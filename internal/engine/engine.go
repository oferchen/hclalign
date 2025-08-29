package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/align"
	"github.com/oferchen/hclalign/internal/diff"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

var (
	testHookAfterParse   func()
	testHookAfterReorder func()
	WriteFileAtomic      = internalfs.WriteFileAtomic
)

// Run is the entrypoint for processing according to cfg. It delegates to
// ProcessReader when reading from STDIN, otherwise it scans targets and runs the
// processing pipeline.
func Run(ctx context.Context, cfg *config.Config) (bool, error) {
	if cfg.Stdin {
		return processReader(ctx, os.Stdin, os.Stdout, cfg)
	}
	files, err := Scan(ctx, cfg)
	if err != nil {
		return false, err
	}
	changed, outs, err := Pipeline(ctx, files, cfg)
	if err != nil {
		return changed, err
	}
	for _, f := range files {
		if out, ok := outs[f]; ok && len(out) > 0 {
			if _, err := fmt.Fprintf(os.Stdout, "\n--- %s ---\n", f); err != nil {
				return changed, err
			}
			if _, err := os.Stdout.Write(out); err != nil {
				return changed, err
			}
		}
	}
	return changed, nil
}

func processReader(ctx context.Context, r io.Reader, w io.Writer, cfg *config.Config) (bool, error) {
	if w == nil {
		w = os.Stdout
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}

	hints := internalfs.DetectHintsFromBytes(data)
	if len(hints.BOM()) > 0 {
		data = data[len(hints.BOM()):]
	}

	// Phase A: format the input prior to alignment.
	formatted := hclwrite.Format(data)
	file, diags := hclwrite.ParseConfig(formatted, "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("parsing error: %v", diags.Errs())
	}
	if err := align.Apply(file, cfg); err != nil {
		return false, err
	}

	formatted = bytes.ReplaceAll(file.Bytes(), []byte("\r\n"), []byte("\n"))
	styled := internalfs.ApplyHints(formatted, hints)
	original := data
	if bom := hints.BOM(); len(bom) > 0 {
		original = append(append([]byte{}, bom...), original...)
	}
	changed := !bytes.Equal(original, styled)

	switch cfg.Mode {
	case config.ModeDiff:
		if changed {
			text, err := diff.Unified("stdin", "stdin", original, styled, hints.Newline)
			if err != nil {
				return false, err
			}
			if _, err := fmt.Fprint(w, text); err != nil {
				return false, err
			}
		}
	default:
		if cfg.Stdout {
			if _, err := w.Write(styled); err != nil {
				return changed, err
			}
		}
	}
	return changed, nil
}
