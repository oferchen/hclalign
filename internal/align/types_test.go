// internal/align/types_test.go
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
	outPath := filepath.Join(base, "out.tf")
	src, err := os.ReadFile(inPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	outBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	outFile, diags := hclwrite.ParseConfig(outBytes, outPath, hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse out: %v", diags)
	}
	cases := []struct {
		name  string
		types map[string]struct{}
		want  func() []byte
	}{
		{
			name:  "variable",
			types: map[string]struct{}{"variable": {}},
			want: func() []byte {
				return []byte("variable \"a\" {\n  description = \"d\"\n  type        = string\n}\n\noutput \"o\" {\n  value       = \"v\"\n  description = \"d\"\n}\n")
			},
		},
		{
			name:  "output",
			types: map[string]struct{}{"output": {}},
			want: func() []byte {
				return []byte("variable \"a\" {\n  description = \"d\"\n  type        = string\n}\n\noutput \"o\" {\n  description = \"d\"\n  value       = \"v\"\n}\n")
			},
		},
		{
			name:  "all",
			types: nil,
			want:  func() []byte { return hclwrite.Format(outFile.Bytes()) },
		},
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
			want := tc.want()
			if !bytes.Equal(got, want) {
				t.Fatalf("output mismatch for %s:\n-- got --\n%s\n-- want --\n%s", tc.name, got, want)
			}
			file2, diags := hclwrite.ParseConfig(want, tc.name, hcl.InitialPos)
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
