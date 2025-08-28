// internal/align/fuzz_test.go
package align

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/internal/hclalign"
)

func FuzzReorderStability(f *testing.F) {
	f.Add(uint64(0))
	f.Fuzz(func(t *testing.T, seed uint64) {
		r := rand.New(rand.NewSource(int64(seed)))
		attrs := []string{"description", "type", "default", "sensitive", "nullable", "extra", "foo"}
		r.Shuffle(len(attrs), func(i, j int) { attrs[i], attrs[j] = attrs[j], attrs[i] })

		const maxFuzzBytes = 1 << 12
		var src bytes.Buffer
		src.WriteString("variable \"fuzz\" {\n")
		for i, name := range attrs {
			addRandomPadding(&src, r)
			fmt.Fprintf(&src, "  %s = %d\n", name, i)
			addRandomPadding(&src, r)
		}
		src.WriteString("}\n")

		if src.Len() > maxFuzzBytes {
			t.Skip("input too large")
		}

		file, diags := hclwrite.ParseConfig(src.Bytes(), "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse: %v", diags)
		}
		hclalign.ReorderAttributes(file, nil, false)
		out := file.Bytes()

		if len(out) > maxFuzzBytes {
			t.Skip("output too large")
		}

		file2, diags := hclwrite.ParseConfig(out, "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse reordered: %v", diags)
		}
		hclalign.ReorderAttributes(file2, nil, false)
		if !bytes.Equal(out, file2.Bytes()) {
			t.Fatalf("round-trip mismatch")
		}
	})
}

func addRandomPadding(buf *bytes.Buffer, r *rand.Rand) {
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
