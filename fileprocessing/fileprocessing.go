// fileprocessing.go
// Handles processing of HCL files based on configuration and operation mode.

package fileprocessing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/sync/semaphore"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/hclprocessing"
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(int64(cfg.Concurrency))
	errChan := make(chan error, cfg.Concurrency)
	changedChan := make(chan bool, cfg.Concurrency)
	var firstErr error
	var once sync.Once

	walkErr := filepath.WalkDir(cfg.Target, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			once.Do(func() { firstErr = err })
			errChan <- err
			cancel()
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}
		if !patternmatching.MatchesFileCriteria(filePath, cfg.Include) || patternmatching.MatchesFileCriteria(filePath, cfg.Exclude) {
			return nil
		}
		if err := sem.Acquire(ctx, 1); err != nil {
			once.Do(func() { firstErr = err })
			errChan <- err
			cancel()
			return err
		}
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			defer sem.Release(1)
			ch, err := processSingleFile(ctx, path, cfg)
			if err != nil {
				once.Do(func() { firstErr = err })
				errChan <- err
				cancel()
				return
			}
			if ch {
				changedChan <- true
			}
		}(filePath)
		return nil
	})

	wg.Wait()
	cancel()
	close(errChan)
	close(changedChan)

	changed := false
	for range changedChan {
		changed = true
	}

	var errs []error
	if walkErr != nil {
		errs = append(errs, walkErr)
	}
	for e := range errChan {
		errs = append(errs, e)
	}
	if firstErr != nil {
		if len(errs) > 1 {
			return changed, errors.Join(errs...)
		}
		return changed, firstErr
	}
	if len(errs) > 0 {
		return changed, errors.Join(errs...)
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

	file, diags := hclwrite.ParseConfig(original, filePath, hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("parsing error in file %s: %v", filePath, diags.Errs())
	}

	hclprocessing.ReorderAttributes(file, cfg.Order)

	formatted := file.Bytes()
	changed := !bytes.Equal(original, formatted)

	switch cfg.Mode {
	case config.ModeWrite:
		if !changed {
			if cfg.Stdout {
				_, _ = os.Stdout.Write(formatted)
			}
			return false, nil
		}
		if err := writeFileAtomically(filePath, formatted, fileInfo.Mode()); err != nil {
			return false, fmt.Errorf("error writing file %s with original permissions: %w", filePath, err)
		}
		if cfg.Stdout {
			_, _ = os.Stdout.Write(formatted)
		}
	case config.ModeCheck:
		if cfg.Stdout {
			_, _ = os.Stdout.Write(formatted)
		}
	case config.ModeDiff:
		if changed {
			ud := difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(original)),
				B:        difflib.SplitLines(string(formatted)),
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

func writeFileAtomically(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	tmp, err := os.CreateTemp(dir, "hclalign-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, filename); err != nil {
		return err
	}
	dirf, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer dirf.Close()
	return dirf.Sync()
}
