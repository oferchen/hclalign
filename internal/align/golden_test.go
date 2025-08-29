// internal/align/golden_test.go
package align

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestGolden(t *testing.T) {
	casesDir := filepath.Join("..", "..", "tests", "cases")
	err := filepath.WalkDir(casesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		inPath := filepath.Join(path, "in.tf")
		outPath := filepath.Join(path, "out.tf")
		if _, err := os.Stat(inPath); err != nil {
			return nil
		}
		if _, err := os.Stat(outPath); err != nil {
			return nil
		}
		strictPath := filepath.Join(path, "out_strict.tf")
		name, err := filepath.Rel(casesDir, path)
		if err != nil {
			return err
		}

		t.Run(name, func(t *testing.T) {
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
				if err := Apply(file, &Options{Strict: strict}); err != nil {
					if strict {
						t.Skipf("strict mode error: %v", err)
						return
					}
					t.Fatalf("reorder: %v", err)
				}
				got := file.Bytes()
				if !bytes.Equal(got, expBytes) {
					t.Fatalf("output mismatch for %s (strict=%v):\n-- got --\n%s\n-- want --\n%s", name, strict, got, expBytes)
				}

				file2, diags := hclwrite.ParseConfig(expBytes, expPath, hcl.InitialPos)
				if diags.HasErrors() {
					t.Fatalf("parse expected: %v", diags)
				}
				if err := Apply(file2, &Options{Strict: strict}); err != nil {
					if strict {
						t.Skipf("strict mode error: %v", err)
						return
					}
					t.Fatalf("reorder expected: %v", err)
				}
				if !bytes.Equal(expBytes, file2.Bytes()) {
					t.Fatalf("non-idempotent on expected for %s (strict=%v)", name, strict)
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
		return nil
	})
	if err != nil {
		t.Fatalf("walk cases: %v", err)
	}
}

func TestUnknownAttributesAfterCanonical(t *testing.T) {
	src := []byte(`variable "example" {
  foo = "foo"
  description = "example"
  bar = "bar"
  type = number
  default = 1
}`)

	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse input: %v", diags)
	}

	if err := Apply(file, &Options{}); err != nil {
		t.Fatalf("reorder: %v", err)
	}

	got := file.Bytes()
	exp := []byte(`variable "example" {
  description = "example"
  type        = number
  default     = 1
  foo         = "foo"
  bar         = "bar"
}`)
	if !bytes.Equal(got, exp) {
		t.Fatalf("output mismatch:\n-- got --\n%s\n-- want --\n%s", got, exp)
	}
}
