package diff

import "github.com/pmezard/go-difflib/difflib"

// Unified generates a unified diff for two byte slices using the given
// file names and line ending. The diff always uses a context of 3 lines.
func Unified(fromFile, toFile string, original, styled []byte, eol string) (string, error) {
	ud := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(original)),
		B:        difflib.SplitLines(string(styled)),
		FromFile: fromFile,
		ToFile:   toFile,
		Context:  3,
		Eol:      eol,
	}
	return difflib.GetUnifiedDiffString(ud)
}
