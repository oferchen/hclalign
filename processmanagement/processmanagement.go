// processmanagement.go
// Orchestrates the reading, processing, and writing of HCL files across different modules.

package processmanagement

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"hclalign/config"
	"hclalign/fileprocessing"
	"hclalign/hclprocessing"
	"hclalign/patternmatching"
)

var (
	errMutex   sync.Mutex // Protects access to the firstError variable
	firstError error      // Stores the first error encountered for reporting
)

func ProcessTargetDynamically(target string, masks []string, order []string) error {
	if len(masks) == 0 {
		return config.FormatGenericError("masks input", fmt.Errorf("cannot be empty"))
	}
	if strings.TrimSpace(target) == "" {
		return config.FormatGenericError("target input", fmt.Errorf("cannot be empty"))
	}

	compiledMasks, err := patternmatching.CompilePatterns(masks)
	if err != nil {
		return config.FormatGenericError("compiling masks", err)
	}

	return processFilesConcurrently(target, compiledMasks, order)
}

func processFilesConcurrently(target string, compiledMasks []*regexp.Regexp, order []string) error {
	var wg sync.WaitGroup
	var firstError error
	errMutex := &sync.Mutex{}

	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !patternmatching.MatchesFileMasks(filepath.Base(path), compiledMasks) {
			return nil // Skip directories and non-matching files.
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			content, readErr := fileprocessing.ReadFile(path)
			if readErr != nil {
				recordError(readErr)
				return
			}

			errMutex.Lock()
			processedContent, processErr := hclprocessing.ProcessContent(content, order)
			if processErr != nil {
				recordError(processErr)
				return
			}

			if writeErr := fileprocessing.WriteFile(path, processedContent, info.Mode()); writeErr != nil {
				recordError(writeErr)
			}
		}()
		return nil
	})

	wg.Wait()
	if err != nil {
		return err // Return if any errors occurred during file walking.
	}
	if firstError != nil {
		return firstError // Return the first error encountered during processing.
	}
	return nil
}

func recordError(err error) {
	errMutex.Lock()
	if firstError == nil {
		firstError = err // Capture the first error encountered.
	}
	errMutex.Unlock()
}
