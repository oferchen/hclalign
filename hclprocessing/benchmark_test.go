// hclprocessing/benchmark_test.go
package hclprocessing

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"testing"
)

func BenchmarkReorderAttributes(b *testing.B) {
	src := []byte(`variable "example" {
  description = "d"
  type        = string
  default     = "v"
  sensitive   = true
  nullable    = false
  extra       = 1
}`)
	for i := 0; i < b.N; i++ {
		f, diags := hclwrite.ParseConfig(src, "test.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			b.Fatalf("parse: %v", diags)
		}
		if err := ReorderAttributes(f, nil, false); err != nil {
			b.Fatalf("reorder: %v", err)
		}
	}
}
