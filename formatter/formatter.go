// formatter/formatter.go
package formatter

import (
        "bytes"
        "fmt"
        "unicode/utf8"

        "github.com/hashicorp/hcl/v2"
        "github.com/hashicorp/hcl/v2/hclwrite"

        internalfs "github.com/oferchen/hclalign/internal/fs"
)


func Format(src []byte, filename string) (formatted []byte, hints internalfs.Hints, err error) {
	hints = internalfs.DetectHintsFromBytes(src)
	if bom := hints.BOM(); len(bom) > 0 {
		src = src[len(bom):]
	}
	if len(src) > 0 && !utf8.Valid(src) {
		err = fmt.Errorf("input is not valid UTF-8")
		return
	}

	f, diags := hclwrite.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		err = diags
		return
	}
	formatted = hclwrite.Format(f.Bytes())
	formatted = bytes.ReplaceAll(formatted, []byte("\r\n"), []byte("\n"))

	if len(formatted) > 0 {
		formatted = bytes.TrimRight(formatted, "\n")
		formatted = append(formatted, '\n')
	}
	return

func Format(src []byte, filename string) ([]byte, internalfs.Hints, error) {
        hints := internalfs.DetectHintsFromBytes(src)
        if bom := hints.BOM(); len(bom) > 0 {
                src = src[len(bom):]
        }
        if len(src) > 0 && !utf8.Valid(src) {
                return nil, internalfs.Hints{}, fmt.Errorf("input is not valid UTF-8")
        }

        f, diags := hclwrite.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
        if diags.HasErrors() {
                return nil, internalfs.Hints{}, diags
        }
        formatted := hclwrite.Format(f.Bytes())
        formatted = bytes.ReplaceAll(formatted, []byte("\r\n"), []byte("\n"))

        if len(formatted) > 0 {
                formatted = bytes.TrimRight(formatted, "\n")
                formatted = append(formatted, '\n')
        }

        return formatted, hints, nil
}
