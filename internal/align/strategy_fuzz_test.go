// internal/align/strategy_fuzz_test.go
package align

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func FuzzStrategies(f *testing.F) {
	f.Add("variable")
	f.Fuzz(func(t *testing.T, typ string) {
		src := new(bytes.Buffer)
		fmt.Fprintf(src, "%s \"x\" {\n", typ)
		fmt.Fprintf(src, "  c = 3\n  a = 1\n  b = 2\n}\n")
		file, diags := hclwrite.ParseConfig(src.Bytes(), "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Skip()
		}
		if err := Apply(file, &Options{}); err != nil {
			t.Skip()
		}
		out := file.Bytes()
		file2, diags := hclwrite.ParseConfig(out, "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse: %v", diags)
		}
		if err := Apply(file2, &Options{}); err != nil {
			t.Fatalf("apply: %v", err)
		}
		if !bytes.Equal(out, file2.Bytes()) {
			t.Fatalf("not idempotent")
		}
	})
}
