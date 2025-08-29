// /internal/align/strict_error_test.go
package align

import (
	"strings"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestStrictOrderErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "missing",
			src:  "variable \"a\" {\n  type = string\n}",
			want: "missing attributes",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, diags := hclwrite.ParseConfig([]byte(tc.src), "test.hcl", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse: %v", diags)
			}
			err := Apply(f, &Options{Strict: true})
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
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "variable",
			src:  "variable \"a\" {\n  description = \"desc\"\n  type = string\n  default = 1\n  sensitive = true\n  nullable = false\n  foo = 1\n}",
		},
		{
			name: "output",
			src:  "output \"o\" {\n  value = 1\n  foo   = 1\n}",
		},
		{
			name: "module",
			src:  "module \"m\" {\n  source = \"./m\"\n  foo    = 1\n}",
		},
		{
			name: "terraform",
			src:  "terraform {\n  required_version = \"1.0\"\n  required_providers {}\n  backend \"local\" {}\n  cloud {}\n  foo = 1\n}",
		},
		{
			name: "dynamic",
			src:  "dynamic \"d\" {\n  for_each = []\n  content {}\n  foo = 1\n}",
		},
		{
			name: "lifecycle",
			src:  "lifecycle {\n  create_before_destroy = true\n  foo = 1\n}",
		},
		{
			name: "provisioner",
			src:  "provisioner \"local-exec\" {\n  when = \"create\"\n  foo  = 1\n}",
		},
		{
			name: "connection",
			src:  "connection {\n  host = \"h\"\n  foo  = 1\n}",
		},
		{
			name: "connection_nested",
			src:  "provisioner \"local-exec\" {\n  connection {\n    host = \"h\"\n    foo  = 1\n  }\n}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, diags := hclwrite.ParseConfig([]byte(tc.src), "test.hcl", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse: %v", diags)
			}
			if err := Apply(f, &Options{Strict: true}); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}
