package hclprocessing

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func FuzzParseHCL(f *testing.F) {
	f.Add([]byte("variable \"t\" {}"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = hclwrite.ParseConfig(data, "fuzz.hcl", hcl.InitialPos)
	})
}

func FuzzAttributeOrdering(f *testing.F) {
	f.Add([]byte("variable \"r\" { default = 1 description = \"d\" }"))
	f.Fuzz(func(t *testing.T, data []byte) {
		file, diags := hclwrite.ParseConfig(data, "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			return
		}
		ReorderAttributes(file, nil)
	})
}
