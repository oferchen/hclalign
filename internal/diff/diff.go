// /internal/diff/diff.go
package diff

import (
	"strings"

	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/pmezard/go-difflib/difflib"
)

const diffContext = 3

func Unified(fromFile, toFile string, original, styled []byte, hints internalfs.Hints) (string, error) {
	if bom := hints.BOM(); len(bom) > 0 {
		original = append(append([]byte{}, bom...), original...)
		styled = append(append([]byte{}, bom...), styled...)
	}
	ud := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(original)),
		B:        difflib.SplitLines(string(styled)),
		FromFile: fromFile,
		ToFile:   toFile,
		Context:  diffContext,
		Eol:      hints.Newline,
	}
	out, err := difflib.GetUnifiedDiffString(ud)
	if err != nil {
		return "", err
	}
	if hints.Newline == "\r\n" && strings.HasSuffix(out, "\n") && !strings.HasSuffix(out, "\r\n") {
		out = strings.TrimSuffix(out, "\n") + "\r\n"
	}
	return out, nil
}
