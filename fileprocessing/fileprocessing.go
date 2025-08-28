// fileprocessing.go
// Handles processing of HCL files based on configuration criteria and order.

package fileprocessing

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/oferchen/hclalign/hclprocessing"
	"github.com/oferchen/hclalign/patternmatching"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"golang.org/x/sync/semaphore"
)

type contextKey string

// TargetContextKey is the context key used to propagate the target path.
const TargetContextKey contextKey = "target"

// ProcessFiles processes files in the specified target directory according to criteria and order.
func ProcessFiles(ctx context.Context, target string, criteria []string, order []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	maxConcurrency := runtime.GOMAXPROCS(0)
	sem := semaphore.NewWeighted(int64(maxConcurrency))
	errChan := make(chan error, maxConcurrency)
	var firstErr error
	var once sync.Once

	walkErr := filepath.WalkDir(target, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			once.Do(func() { firstErr = err })
			errChan <- err
			cancel()
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !d.IsDir() && patternmatching.MatchesFileCriteria(filePath, criteria) {
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
				if err := ProcessSingleFile(ctx, path, order); err != nil {
					once.Do(func() { firstErr = err })
					errChan <- err
					cancel()
				}
			}(filePath)
		}
		return nil
	})

	wg.Wait()
	cancel()
	close(errChan)

	var errs []error
	if walkErr != nil {
		errs = append(errs, walkErr)
	}
	for e := range errChan {
		errs = append(errs, e)
	}
	if firstErr != nil {
		if len(errs) > 1 {
			return errors.Join(errs...)
		}
		return firstErr
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// ProcessSingleFile reads and processes a single HCL file based on the given order.
func ProcessSingleFile(ctx context.Context, filePath string, order []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error retrieving file info for %s: %w", filePath, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	file, diags := hclwrite.ParseConfig(fileContent, filePath, hcl.InitialPos)
	if diags.HasErrors() {
		return fmt.Errorf("parsing error in file %s: %v", filePath, diags.Errs())
	}

	hclprocessing.ReorderAttributes(file, order)

	if err := writeFileAtomically(filePath, file.Bytes(), fileInfo.Mode()); err != nil {
		return fmt.Errorf("error writing file %s with original permissions: %w", filePath, err)
	}

	return nil
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
