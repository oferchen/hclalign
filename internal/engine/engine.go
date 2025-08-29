// internal/engine/engine.go
package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/align"
	"github.com/oferchen/hclalign/internal/diff"
	terraformfmt "github.com/oferchen/hclalign/internal/fmt"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

var (
	testHookAfterParse   func()
	testHookAfterReorder func()
	WriteFileAtomic      = internalfs.WriteFileAtomic
)

func Process(ctx context.Context, cfg *config.Config) (bool, error) {
	if cfg.Stdin {
		return processReader(ctx, os.Stdin, os.Stdout, cfg)
	}
	return processFiles(ctx, cfg)
}

func processFiles(ctx context.Context, cfg *config.Config) (bool, error) {
	files, err := scan(ctx, cfg)
	if err != nil {
		return false, err
	}
	outs, changed, errs := runPipeline(ctx, cfg, files)

	for i, f := range files {
		if out, ok := outs[f]; ok && len(out) > 0 {
			header := "--- %s ---\n"
			if i == 0 {
				header = "\n--- %s ---\n"
			}
			if _, err := fmt.Fprintf(os.Stdout, header, f); err != nil {
				return changed, err
			}
			if i < len(files)-1 {
				if len(out) == 0 || out[len(out)-1] != '\n' {
					out = append(out, '\n')
				}
			} else if len(files) > 1 && len(out) > 0 && out[len(out)-1] == '\n' {
				out = out[:len(out)-1]
			}
			if _, err := os.Stdout.Write(out); err != nil {
				return changed, err
			}
		}
	}

	if len(errs) > 0 {
		return changed, errors.Join(errs...)
	}
	return changed, nil
}

func processReader(ctx context.Context, r io.Reader, w io.Writer, cfg *config.Config) (bool, error) {
	if w == nil {
		w = os.Stdout
	}

	data, hints, err := internalfs.ReadAllWithHints(r)
	if err != nil {
		return false, err
	}

	original := data
	hadNewline := len(data) > 0 && data[len(data)-1] == '\n'
	formatted := data
	if !cfg.NoFmt {
		formatted, err = terraformfmt.Format(data, "stdin", cfg.FmtStrategy)
		if err != nil {
			return false, fmt.Errorf("parsing error: %w", err)
		}
	}

	parseData := internalfs.PrepareForParse(formatted, hints)

	if !cfg.FmtOnly {
		file, diags := hclwrite.ParseConfig(parseData, "stdin", hcl.InitialPos)
		if diags.HasErrors() {
			return false, fmt.Errorf("parsing error: %v", diags.Errs())
		}
		schemas, err := loadSchemas(ctx, cfg)
		if err != nil {
			return false, err
		}
		if err := align.Apply(file, &align.Options{Order: cfg.Order, Strict: cfg.StrictOrder, Schemas: schemas}); err != nil {
			return false, err
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

	switch cfg.Mode {
	case config.ModeDiff:
		if changed {
			text, err := diff.Unified("stdin", "stdin", original, styled, hints)
			if err != nil {
				return false, err
			}
			if _, err := fmt.Fprint(w, text); err != nil {
				return false, err
			}
		}
	default:
		if cfg.Stdout {
			if err := internalfs.WriteAllWithHints(w, formatted, hints); err != nil {
				return changed, err
			}
		}
	}
	return changed, nil
}
