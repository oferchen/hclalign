// /internal/engine/pipeline.go
package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/align"
	"github.com/oferchen/hclalign/internal/diff"
	terraformfmt "github.com/oferchen/hclalign/internal/fmt"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

func runPipeline(ctx context.Context, cfg *config.Config, files []string) (map[string][]byte, bool, []error) {
	outs := make(map[string][]byte, len(files))

	schemas, err := loadSchemas(ctx, cfg)
	if err != nil {
		return outs, false, []error{err}
	}

	fileCh := make(chan string)
	results := make(chan struct {
		path string
		data []byte
	}, len(files))

	var changed atomic.Bool
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var errs []error

	go func() {
		defer close(fileCh)
		for _, f := range files {
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
					ch, out, err := processFile(ctx, f, cfg, schemas)
					if err != nil {
						if errors.Is(err, context.Canceled) {
							return
						}
						errMu.Lock()
						errs = append(errs, fmt.Errorf("%s: %w", f, err))
						errMu.Unlock()
						continue
					}
					if ch {
						changed.Store(true)
					}
					if len(out) > 0 {
						select {
						case results <- struct {
							path string
							data []byte
						}{path: f, data: out}:
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
		return outs, changed.Load(), errs
	}
	if err := ctx.Err(); err != nil {
		return outs, changed.Load(), []error{err}
	}
	return outs, changed.Load(), nil
}

func processFile(ctx context.Context, filePath string, cfg *config.Config, schemas map[string]*align.Schema) (bool, []byte, error) {
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}
	data, perm, hints, err := internalfs.ReadFileWithHints(ctx, filePath)
	if err != nil {
		return false, nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}

	original := data
	hadNewline := len(data) > 0 && data[len(data)-1] == '\n'
	formatted := data
	if !cfg.NoFmt {
		formatted, err = terraformfmt.Format(data, filePath, cfg.FmtStrategy)
		if err != nil {
			return false, nil, fmt.Errorf("parsing error in file %s: %w", filePath, err)
		}
	}

	parseData := internalfs.PrepareForParse(formatted, hints)

	if !cfg.FmtOnly {
		file, diags := hclwrite.ParseConfig(parseData, filePath, hcl.InitialPos)
		if diags.HasErrors() {
			return false, nil, fmt.Errorf("parsing error in file %s: %v", filePath, diags.Errs())
		}
		if testHookAfterParse != nil {
			testHookAfterParse()
		}
		if err := align.Apply(file, &align.Options{Order: cfg.Order, Strict: cfg.StrictOrder, Schemas: schemas}); err != nil {
			return false, nil, err
		}
		if testHookAfterReorder != nil {
			testHookAfterReorder()
		}
		formatted = hclwrite.Format(file.Bytes())
	}

	if !hadNewline && len(formatted) > 0 && formatted[len(formatted)-1] == '\n' {
		formatted = formatted[:len(formatted)-1]
	}

	styled := internalfs.ApplyHints(formatted, hints)
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
		if err := WriteFileAtomic(ctx, filePath, formatted, perm, hints); err != nil {
			return false, nil, fmt.Errorf("error writing file %s with original permissions: %w", filePath, err)
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
			text, err := diff.Unified(filePath, filePath, original, styled, hints)
			if err != nil {
				return false, nil, err
			}
			out = []byte(text)
		}
	}

	return changed, out, nil
}
