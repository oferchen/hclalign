// /internal/align/types_test.go
package align_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
)

func TestTypesSelection(t *testing.T) {
	base := filepath.Join("..", "..", "tests", "cases", "types")
	inPath := filepath.Join(base, "in.tf")
	src, err := os.ReadFile(inPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	cases := []struct {
		name  string
		types map[string]struct{}
		want  string
	}{
		{"variable", map[string]struct{}{"variable": {}}, "aligned_variable.tf"},
		{"output", map[string]struct{}{"output": {}}, "aligned_output.tf"},
		{"all", nil, "aligned_all.tf"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file, diags := hclwrite.ParseConfig(src, inPath, hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse input: %v", diags)
			}
			if err := alignpkg.Apply(file, &alignpkg.Options{Types: tc.types}); err != nil {
				t.Fatalf("align: %v", err)
			}
			got := hclwrite.Format(file.Bytes())
			want, err := os.ReadFile(filepath.Join(base, tc.want))
			if err != nil {
				t.Fatalf("read expected: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("output mismatch for %s:\n-- got --\n%s\n-- want --\n%s", tc.name, got, want)
			}
			file2, diags := hclwrite.ParseConfig(want, tc.want, hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse expected: %v", diags)
			}
			if err := alignpkg.Apply(file2, &alignpkg.Options{Types: tc.types}); err != nil {
				t.Fatalf("reapply: %v", err)
			}
			if !bytes.Equal(want, hclwrite.Format(file2.Bytes())) {
				t.Fatalf("non-idempotent for %s", tc.name)
			}
		})
	}
}
