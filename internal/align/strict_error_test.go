// internal/align/strict_error_test.go
package align

import (
	"strings"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/internal/hclalign"
)

func TestStrictOrderErrors(t *testing.T) {
	cases := []struct {
		name	string
		src	string
		want	string
	}{
		{
			name:	"missing",
			src:	"variable \"a\" {\n  type = string\n}",
			want:	"missing attributes",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, diags := hclwrite.ParseConfig([]byte(tc.src), "test.hcl", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse: %v", diags)
			}
			err := hclalign.ReorderAttributes(f, nil, true)
			if err == nil {
				t.Fatalf("expected error")
			}
			if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err, tc.want)
			}
		})
	}
}

func TestStrictOrderRejectsUnknownAttributes(t *testing.T) {
	src := "variable \"a\" {\n  description = \"desc\"\n  type = string\n  default = 1\n  sensitive = true\n  nullable = false\n  foo = 1\n}"
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	if err := hclalign.ReorderAttributes(f, nil, true); err == nil {
		t.Fatalf("expected error")
	}
}

