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
	var walk func(string) error
	walk = func(dir string) error {
		if !matcher.Matches(dir) {
			return nil
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			if entry.Type()&os.ModeSymlink != 0 {
				info, err := os.Stat(path)
				if err != nil {
					return err
				}
				if info.IsDir() {
					if cfg.FollowSymlinks {
						if err := walk(path); err != nil {
							return err
						}
					}
					continue
				}
			}
			if entry.IsDir() {
				if err := walk(path); err != nil {
					return err
				}
				continue
			}
			if matcher.Matches(path) {
				files = append(files, path)
			}
		}
		return nil
	}
	info, err := os.Lstat(cfg.Target)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		if err := walk(cfg.Target); err != nil {
			return false, err
		}
	} else {
		if matcher.Matches(cfg.Target) {
			files = append(files, cfg.Target)
		}
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
	data, perm, hints, err := internalfs.ReadFileWithHints(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	file, diags := hclwrite.ParseConfig(data, filePath, hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("parsing error in file %s: %v", filePath, diags.Errs())
	}

	hclprocessing.ReorderAttributes(file, cfg.Order, cfg.StrictOrder)

	formatted := file.Bytes()
	styled := internalfs.ApplyHints(formatted, hints)
	original := internalfs.ApplyHints(data, hints)
	changed := !bytes.Equal(original, styled)

	switch cfg.Mode {
	case config.ModeWrite:
		if !changed {
			if cfg.Stdout {
				_, _ = os.Stdout.Write(styled)
			}
			return false, nil
		}
		if err := internalfs.WriteFileAtomic(filePath, formatted, perm, hints); err != nil {
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

	hints := internalfs.DetectHintsFromBytes(data)
	if len(hints.BOM()) > 0 {
		data = data[len(hints.BOM()):]
	}

	file, diags := hclwrite.ParseConfig(data, "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("parsing error: %v", diags.Errs())
	}
	hclprocessing.ReorderAttributes(file, cfg.Order, cfg.StrictOrder)
	formatted := file.Bytes()
	styled := internalfs.ApplyHints(formatted, hints)
	original := internalfs.ApplyHints(data, hints)
	changed := !bytes.Equal(original, styled)

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
