// filename: internal/diff/diff_bom_test.go
package diff

import (
	"strings"
	"testing"

	internalfs "github.com/oferchen/hclalign/internal/fs"
)

func TestUnifiedDiffWithBOM(t *testing.T) {
	a := []byte("line1\n")
	b := []byte("line2\n")
	diffStr, err := Unified(UnifiedOpts{FromFile: "a", ToFile: "a", Original: a, Styled: b, Hints: internalfs.Hints{HasBOM: true, Newline: "\n"}})
	if err != nil {
		t.Fatalf("Unified returned error: %v", err)
	}
	if !strings.Contains(diffStr, "-\ufeffline1\n+\ufeffline2\n") {
		t.Fatalf("expected BOM in diff, got: %q", diffStr)
	}
}

func TestUnifiedDiffWithBOMCRLF(t *testing.T) {
	a := []byte("line1\r\n")
	b := []byte("line2\r\n")
	diffStr, err := Unified(UnifiedOpts{FromFile: "a", ToFile: "a", Original: a, Styled: b, Hints: internalfs.Hints{HasBOM: true, Newline: "\r\n"}})
	if err != nil {
		t.Fatalf("Unified returned error: %v", err)
	}
	if !strings.HasSuffix(diffStr, "\r\n") {
		t.Fatalf("expected CRLF ending, got: %q", diffStr)
	}
	if !strings.Contains(diffStr, "-\ufeffline1\r\n+\ufeffline2\r\n") {
		t.Fatalf("expected BOM with CRLF, got: %q", diffStr)
	}
}
