// fileprocessing.go
// Handles processing of HCL files based on configuration and operation mode.

package fileprocessing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

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
	changed := false
	for _, f := range files {
		ch, err := processSingleFile(ctx, f, cfg)
		if err != nil {
			return changed, err
		}
		if ch {
			changed = true
		}
	}
	return changed, nil
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
			text, _ := difflib.GetUnifiedDiffString(ud)
			fmt.Fprint(os.Stdout, text)
		}
	}

	return changed, nil
}

func processReader(ctx context.Context, r io.Reader, cfg *config.Config) (bool, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}
	file, diags := hclwrite.ParseConfig(data, "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("parsing error: %v", diags.Errs())
	}
	hclprocessing.ReorderAttributes(file, cfg.Order)
	formatted := file.Bytes()
	changed := !bytes.Equal(data, formatted)

	switch cfg.Mode {
	case config.ModeDiff:
		if changed {
			ud := difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(data)),
				B:        difflib.SplitLines(string(formatted)),
				FromFile: "stdin",
				ToFile:   "stdin",
				Context:  3,
			}
			text, _ := difflib.GetUnifiedDiffString(ud)
			fmt.Fprint(os.Stdout, text)
		}
	default:
		if cfg.Stdout {
			_, _ = os.Stdout.Write(formatted)
		}
	}
	return changed, nil
}
