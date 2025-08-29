// internal/diff/diff_test.go
package diff

import (
	"strings"
	"testing"
)

func TestUnifiedDiff(t *testing.T) {
	a := []byte("line1\nline2\n")
	b := []byte("line1\nline3\n")
	diffStr, err := Unified("a", "a", a, b, "\n")
	if err != nil {
		t.Fatalf("Unified returned error: %v", err)
	}
	expected := strings.Join([]string{
		"--- a",
		"+++ a",
		"@@ -1,3 +1,3 @@",
		" line1",
		"-line2",
		"+line3",
		" ",
	}, "\n") + "\n"
	if diffStr != expected {
		t.Fatalf("unexpected diff:\n%q\nexpected:\n%q", diffStr, expected)
	}
}

func TestUnifiedDiffUsesEOL(t *testing.T) {
	a := []byte("line1\r\nline2\r\n")
	b := []byte("line1\r\nline3\r\n")
	diffStr, err := Unified("a", "a", a, b, "\r\n")
	if err != nil {
		t.Fatalf("Unified returned error: %v", err)
	}
	if !strings.Contains(diffStr, "-line2\r\n+line3\r\n") {
		t.Fatalf("expected CRLF line endings in diff, got: %q", diffStr)
	}
}

