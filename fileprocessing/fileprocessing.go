// fileprocessing.go
// Handles processing of HCL files based on configuration criteria and order.

package fileprocessing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/hclprocessing"
	"github.com/oferchen/hclalign/patternmatching"
	"golang.org/x/sync/semaphore"
)

// ProcessFiles processes files in the specified target directory according to criteria and order.
func ProcessFiles(target string, criteria []string, order []string) error {
	compiledPatterns, err := patternmatching.CompilePatterns(criteria)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	maxConcurrency := runtime.GOMAXPROCS(0)
	sem := semaphore.NewWeighted(int64(maxConcurrency))

	ctx := context.Background()

	err = filepath.WalkDir(target, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && patternmatching.MatchesFileCriteria(filePath, compiledPatterns) {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer sem.Release(1)

				if err := processSingleFile(filePath, order); err != nil {
					select {
					case errChan <- err:
					default:
					}
				}
			}()
		}
		return nil
	})

	wg.Wait()
	close(errChan)
	if err == nil {
		err, _ = <-errChan
	}
	return err
}

// processSingleFile reads and processes a single HCL file based on the given order.
func processSingleFile(filePath string, order []string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	file, diags := hclwrite.ParseConfig(fileContent, filePath, hcl.InitialPos)
	if diags.HasErrors() {
		return fmt.Errorf("parsing error in file %s: %v", filePath, diags.Errs())
	}

	hclprocessing.ReorderAttributes(file, order)

	if err := os.WriteFile(filePath, file.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", filePath, err)
	}

	return nil
}
