package align

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/hclprocessing"
)

// TestGolden runs alignment tests over test cases in tests/cases/<name>.
func TestGolden(t *testing.T) {
	casesDir := filepath.Join("..", "..", "tests", "cases")
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("read cases: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			inPath := filepath.Join(casesDir, name, "in.tf")
			outPath := filepath.Join(casesDir, name, "out.tf")

			inBytes, err := os.ReadFile(inPath)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}
			expBytes, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("read expected: %v", err)
			}

			file, diags := hclwrite.ParseConfig(inBytes, inPath, hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse input: %v", diags)
			}
			hclprocessing.ReorderAttributes(file, nil)
			got := file.Bytes()
			if !bytes.Equal(got, expBytes) {
				t.Fatalf("output mismatch for %s:\n-- got --\n%s\n-- want --\n%s", name, got, expBytes)
			}

			// Idempotency check is required for the stress case.
			if name == "stress" {
				file2, diags := hclwrite.ParseConfig(expBytes, outPath, hcl.InitialPos)
				if diags.HasErrors() {
					t.Fatalf("parse expected: %v", diags)
				}
				hclprocessing.ReorderAttributes(file2, nil)
				if !bytes.Equal(expBytes, file2.Bytes()) {
					t.Fatalf("non-idempotent on expected for %s", name)
				}
			}
		})
	}
}
