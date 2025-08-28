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

// FuzzReorder generates random attribute orders with optional padding and
// ensures reorder is stable.
func FuzzReorder(f *testing.F) {
	f.Add(uint64(0))
	f.Fuzz(func(t *testing.T, seed uint64) {
		r := rand.New(rand.NewSource(int64(seed)))
		attrs := []string{"description", "type", "default", "sensitive", "nullable", "extra", "foo"}
		r.Shuffle(len(attrs), func(i, j int) { attrs[i], attrs[j] = attrs[j], attrs[i] })

		const maxSize = 1 << 12 // 4KiB limit to avoid resource exhaustion
		var src bytes.Buffer
		src.WriteString("variable \"fuzz\" {\n")
		for i, name := range attrs {
			insertPadding(&src, r)
			fmt.Fprintf(&src, "  %s = %d\n", name, i)
			insertPadding(&src, r)
		}
		src.WriteString("}\n")

		if src.Len() > maxSize {
			t.Skip("input too large")
		}

		file, diags := hclwrite.ParseConfig(src.Bytes(), "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse: %v", diags)
		}
		hclprocessing.ReorderAttributes(file, nil, false)
		out := file.Bytes()

		if len(out) > maxSize {
			t.Skip("output too large")
		}

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

// insertPadding randomly adds comments or blank lines around attributes.
func insertPadding(buf *bytes.Buffer, r *rand.Rand) {
	for j := 0; j < r.Intn(3); j++ {
		switch r.Intn(3) {
		case 0:
			buf.WriteByte('\n')
		case 1:
			fmt.Fprintf(buf, "  // c%d\n", r.Intn(100))
		default:
			fmt.Fprintf(buf, "  # c%d\n", r.Intn(100))
		}
	}
}
