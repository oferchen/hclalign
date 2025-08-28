package align

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/hclprocessing"
)

// FuzzReorder generates random attribute orders and ensures reorder is stable.
func FuzzReorder(f *testing.F) {
	f.Add(uint64(0))
	f.Fuzz(func(t *testing.T, seed uint64) {
		attrs := []string{"description", "type", "default", "sensitive", "nullable", "extra", "foo"}
		r := rand.New(rand.NewSource(int64(seed)))
		r.Shuffle(len(attrs), func(i, j int) { attrs[i], attrs[j] = attrs[j], attrs[i] })

		var src bytes.Buffer
		src.WriteString("variable \"fuzz\" {\n")
		for i, name := range attrs {
			fmt.Fprintf(&src, "  %s = %d\n", name, i)
		}
		src.WriteString("}\n")

		file, diags := hclwrite.ParseConfig(src.Bytes(), "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse: %v", diags)
		}
		hclprocessing.ReorderAttributes(file, nil, false)
		out := file.Bytes()

		file2, diags := hclwrite.ParseConfig(out, "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse reordered: %v", diags)
		}
		hclprocessing.ReorderAttributes(file2, nil, false)
		if !bytes.Equal(out, file2.Bytes()) {
			t.Fatalf("round-trip mismatch")
		}
	})
}
