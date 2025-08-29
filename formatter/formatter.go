// /formatter/formatter.go
package formatter

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	internalfs "github.com/oferchen/hclalign/internal/fs"
)

func Format(src []byte, filename string) ([]byte, error) {
	hints := internalfs.DetectHintsFromBytes(src)
	if bom := hints.BOM(); len(bom) > 0 {
		src = src[len(bom):]
	}
	if len(src) > 0 && !utf8.Valid(src) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	f, diags := hclwrite.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	formatted := hclwrite.Format(f.Bytes())
	formatted = bytes.ReplaceAll(formatted, []byte("\r\n"), []byte("\n"))

	if len(formatted) > 0 {
		formatted = bytes.TrimRight(formatted, "\n")
		formatted = append(formatted, '\n')
	}

	formatted = internalfs.ApplyHints(formatted, hints)
	return formatted, nil
}
