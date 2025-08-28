// fileprocessing.go
// Handles processing of HCL files based on configuration and operation mode.

package fileprocessing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/pmezard/go-difflib/difflib"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/hclprocessing"
	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/oferchen/hclalign/patternmatching"
)

// Process processes files or stdin based on the provided configuration. It
// returns a boolean indicating whether any changes were required.
func Process(ctx context.Context, cfg *config.Config) (bool, error) {
	if cfg.Stdin {
		return processReader(ctx, os.Stdin, cfg)
	}
	return processFiles(ctx, cfg)
}

func processFiles(ctx context.Context, cfg *config.Config) (bool, error) {
	matcher, err := patternmatching.NewMatcher(cfg.Include, cfg.Exclude)
	if err != nil {
		return false, err
	}
	var files []string
	err = filepath.WalkDir(cfg.Target, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if !matcher.Matches(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if matcher.Matches(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	sort.Strings(files)

	sem := make(chan struct{}, cfg.Concurrency)
	g, ctx := errgroup.WithContext(ctx)
	var changed atomic.Bool
	for _, f := range files {
		f := f
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return changed.Load(), ctx.Err()
		}
		g.Go(func() error {
			defer func() { <-sem }()
			ch, err := processSingleFile(ctx, f, cfg)
			if err != nil {
				return err
			}
			if ch {
				changed.Store(true)
			}
			if cfg.Verbose {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					log.Printf("processed file: %s", f)
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return changed.Load(), err
	}
	return changed.Load(), nil
}

func processSingleFile(ctx context.Context, filePath string, cfg *config.Config) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, fmt.Errorf("error retrieving file info for %s: %w", filePath, err)
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	original, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	bom := []byte{}
	content := original
	if bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}) {
		bom = []byte{0xEF, 0xBB, 0xBF}
		content = content[len(bom):]
	}
	newline := []byte("\n")
	if bytes.Contains(content, []byte("\r\n")) {
		newline = []byte("\r\n")
	}

	file, diags := hclwrite.ParseConfig(content, filePath, hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("parsing error in file %s: %v", filePath, diags.Errs())
	}

	hclprocessing.ReorderAttributes(file, cfg.Order)

	formatted := file.Bytes()
	styled := internalfs.ApplyHints(formatted, newline, bom)
	changed := !bytes.Equal(original, styled)

	switch cfg.Mode {
	case config.ModeWrite:
		if !changed {
			if cfg.Stdout {
				_, _ = os.Stdout.Write(styled)
			}
			return false, nil
		}
		if err := internalfs.WriteFile(ctx, filePath, formatted, fileInfo.Mode(), newline, bom); err != nil {
			return false, fmt.Errorf("error writing file %s with original permissions: %w", filePath, err)
		}
		if cfg.Stdout {
			_, _ = os.Stdout.Write(styled)
		}
	case config.ModeCheck:
		if cfg.Stdout {
			_, _ = os.Stdout.Write(styled)
		}
	case config.ModeDiff:
		if changed {
			ud := difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(original)),
				B:        difflib.SplitLines(string(styled)),
				FromFile: filePath,
				ToFile:   filePath,
				Context:  3,
			}
			text, err := difflib.GetUnifiedDiffString(ud)
			if err != nil {
				return false, err
			}
			if _, err := fmt.Fprint(os.Stdout, text); err != nil {
				return false, err
			}
		}
	}

	return changed, nil
}

func processReader(ctx context.Context, r io.Reader, cfg *config.Config) (bool, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}

	bom := []byte{}
	content := data
	if bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}) {
		bom = []byte{0xEF, 0xBB, 0xBF}
		content = content[len(bom):]
	}
	newline := []byte("\n")
	if bytes.Contains(content, []byte("\r\n")) {
		newline = []byte("\r\n")
	}

	file, diags := hclwrite.ParseConfig(content, "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("parsing error: %v", diags.Errs())
	}
	hclprocessing.ReorderAttributes(file, cfg.Order)
	formatted := file.Bytes()
	styled := internalfs.ApplyHints(formatted, newline, bom)
	changed := !bytes.Equal(data, styled)

	switch cfg.Mode {
	case config.ModeDiff:
		if changed {
			ud := difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(data)),
				B:        difflib.SplitLines(string(styled)),
				FromFile: "stdin",
				ToFile:   "stdin",
				Context:  3,
			}
			text, err := difflib.GetUnifiedDiffString(ud)
			if err != nil {
				return false, err
			}
			if _, err := fmt.Fprint(os.Stdout, text); err != nil {
				return false, err
			}
		}
	default:
		if cfg.Stdout {
			_, _ = os.Stdout.Write(styled)
		}
	}
	return changed, nil
}
