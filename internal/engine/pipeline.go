package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/align"
	"github.com/oferchen/hclalign/internal/diff"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

// Pipeline executes the formatting and alignment phases for the provided files
// using a worker pool bounded by cfg.Concurrency. It returns whether any file
// changed and a map of file path to output bytes (for stdout/diff modes).
func Pipeline(ctx context.Context, files []string, cfg *config.Config) (bool, map[string][]byte, error) {
	type result struct {
		path string
		data []byte
	}

	var changed atomic.Bool
	outs := make(map[string][]byte, len(files))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fileCh := make(chan string)
	results := make(chan result, len(files))

	var wg sync.WaitGroup
	var errMu sync.Mutex
	var errs []error

	go func() {
		defer close(fileCh)
		for _, f := range files {
			if ctx.Err() != nil {
				return
			}
			select {
			case fileCh <- f:
			case <-ctx.Done():
				return
			}
		}
	}()

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case f, ok := <-fileCh:
					if !ok {
						return
					}
					ch, out, err := processFile(ctx, f, cfg)
					if err != nil {
						if !errors.Is(err, context.Canceled) {
							log.Printf("error processing file %s: %v", f, err)
							errMu.Lock()
							errs = append(errs, fmt.Errorf("%s: %w", f, err))
							errMu.Unlock()
							cancel()
						}
						return
					}
					if ch {
						changed.Store(true)
					}
					if len(out) > 0 {
						select {
						case results <- result{path: f, data: out}:
						case <-ctx.Done():
							return
						}
					}
					if cfg.Verbose {
						log.Printf("processed file: %s", f)
					}
				}
			}
		}()
	}

	wg.Wait()
	close(results)
	for r := range results {
		outs[r.path] = r.data
	}

	if len(errs) > 0 {
		messages := make([]string, len(errs))
		for i, err := range errs {
			messages[i] = err.Error()
		}
		return changed.Load(), nil, errors.New(strings.Join(messages, "\n"))
	}
	if err := ctx.Err(); err != nil {
		return changed.Load(), nil, err
	}

	return changed.Load(), outs, nil
}

func processFile(ctx context.Context, path string, cfg *config.Config) (bool, []byte, error) {
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}
	data, perm, hints, err := internalfs.ReadFileWithHints(ctx, path)
	if err != nil {
		return false, nil, fmt.Errorf("error reading file %s: %w", path, err)
	}
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}

	// Phase A: format the file using hclwrite formatter.
	formatted := hclwrite.Format(data)

	file, diags := hclwrite.ParseConfig(formatted, path, hcl.InitialPos)
	if diags.HasErrors() {
		return false, nil, fmt.Errorf("parsing error in file %s: %v", path, diags.Errs())
	}
	if testHookAfterParse != nil {
		testHookAfterParse()
	}
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}

	// Phase B: apply alignment strategies.
	if err := align.Apply(file, cfg); err != nil {
		return false, nil, err
	}
	if testHookAfterReorder != nil {
		testHookAfterReorder()
	}
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}

	formatted = bytes.ReplaceAll(file.Bytes(), []byte("\r\n"), []byte("\n"))
	styled := internalfs.ApplyHints(formatted, hints)
	original := data
	if bom := hints.BOM(); len(bom) > 0 {
		original = append(append([]byte{}, bom...), original...)
	}
	changed := !bytes.Equal(original, styled)

	var out []byte
	switch cfg.Mode {
	case config.ModeWrite:
		if !changed {
			if cfg.Stdout {
				out = styled
			}
			return false, out, nil
		}
		if err := WriteFileAtomic(ctx, path, formatted, perm, hints); err != nil {
			return false, nil, fmt.Errorf("error writing file %s with original permissions: %w", path, err)
		}
		if cfg.Stdout {
			out = styled
		}
	case config.ModeCheck:
		if cfg.Stdout {
			out = styled
		}
	case config.ModeDiff:
		if changed {
			text, err := diff.Unified(path, path, original, styled, hints.Newline)
			if err != nil {
				return false, nil, err
			}
			out = []byte(text)
		}
	}
	return changed, out, nil
}
