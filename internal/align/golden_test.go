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
			strictPath := filepath.Join(casesDir, name, "out_strict.tf")

			inBytes, err := os.ReadFile(inPath)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}

			run := func(t *testing.T, strict bool, expPath string) {
				expBytes, err := os.ReadFile(expPath)
				if err != nil {
					t.Fatalf("read expected: %v", err)
				}

				file, diags := hclwrite.ParseConfig(inBytes, inPath, hcl.InitialPos)
				if diags.HasErrors() {
					t.Fatalf("parse input: %v", diags)
				}
				hclprocessing.ReorderAttributes(file, nil, strict)
				got := file.Bytes()
				if !bytes.Equal(got, expBytes) {
					t.Fatalf("output mismatch for %s (strict=%v):\n-- got --\n%s\n-- want --\n%s", name, strict, got, expBytes)
				}

				if name == "stress" {
					file2, diags := hclwrite.ParseConfig(expBytes, expPath, hcl.InitialPos)
					if diags.HasErrors() {
						t.Fatalf("parse expected: %v", diags)
					}
					hclprocessing.ReorderAttributes(file2, nil, strict)
					if !bytes.Equal(expBytes, file2.Bytes()) {
						t.Fatalf("non-idempotent on expected for %s (strict=%v)", name, strict)
					}
				}
			}

			t.Run("loose", func(t *testing.T) { run(t, false, outPath) })

			if _, err := os.Stat(strictPath); err == nil {
				t.Run("strict", func(t *testing.T) { run(t, true, strictPath) })
			} else if os.IsNotExist(err) {
				t.Run("strict", func(t *testing.T) { run(t, true, outPath) })
			} else if err != nil {
				t.Fatalf("stat strict: %v", err)
			}
		})
	}
}
