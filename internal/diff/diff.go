// internal/diff/diff.go
package diff

import "github.com/pmezard/go-difflib/difflib"

const diffContext = 3

func Unified(fromFile, toFile string, original, styled []byte, eol string) (string, error) {
	ud := difflib.UnifiedDiff{
		A:		difflib.SplitLines(string(original)),
		B:		difflib.SplitLines(string(styled)),
		FromFile:	fromFile,
		ToFile:		toFile,
		Context:	diffContext,
		Eol:		eol,
	}
	return difflib.GetUnifiedDiffString(ud)
}

