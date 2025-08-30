// filename: internal/align/golden_test.go
package align_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	alignschema "github.com/oferchen/hclalign/internal/align/schema"
	terraformfmt "github.com/oferchen/hclalign/internal/fmt"
)

func TestGolden(t *testing.T) {
	casesDir := filepath.Join("..", "..", "tests", "cases")
	allowed := map[string]struct{}{
		"module":    {},
		"provider":  {},
		"terraform": {},
		"resource":  {},
		"data":      {},
		"variable":  {},
	}
	schemaPath := filepath.Join("..", "..", "tests", "testdata", "providers-schema.json")
	schemas, err := alignschema.LoadFile(schemaPath)
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}
	err = filepath.WalkDir(casesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		inPath := filepath.Join(path, "in.tf")
		fmtPath := filepath.Join(path, "fmt.tf")
		alignedPath := filepath.Join(path, "aligned.tf")
		if _, err := os.Stat(inPath); err != nil {
			return nil
		}
		if _, err := os.Stat(fmtPath); err != nil {
			return nil
		}
		if _, err := os.Stat(alignedPath); err != nil {
			return nil
		}
		name, err := filepath.Rel(casesDir, path)
		if err != nil {
			return err
		}
		if _, ok := allowed[name]; !ok {
			return nil
		}

		t.Run(name, func(t *testing.T) {
			inBytes, err := os.ReadFile(inPath)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}
			fmtBytes, err := os.ReadFile(fmtPath)
			if err != nil {
				t.Fatalf("read fmt: %v", err)
			}
			alignedBytes, err := os.ReadFile(alignedPath)
			if err != nil {
				t.Fatalf("read aligned: %v", err)
			}

			gotFmt, err := terraformfmt.Format(inBytes, inPath, string(terraformfmt.StrategyGo))
			if err != nil {
				t.Fatalf("format input: %v", err)
			}
			if !bytes.Equal(gotFmt, fmtBytes) {
				t.Fatalf("fmt mismatch for %s:\n-- got --\n%s\n-- want --\n%s", name, gotFmt, fmtBytes)
			}
			againFmt, err := terraformfmt.Format(fmtBytes, fmtPath, string(terraformfmt.StrategyGo))
			if err != nil {
				t.Fatalf("format fmt: %v", err)
			}
			if !bytes.Equal(againFmt, fmtBytes) {
				t.Fatalf("fmt not idempotent for %s", name)
			}

			file, diags := hclwrite.ParseConfig(fmtBytes, fmtPath, hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse fmt: %v", diags)
			}
			opts := &alignpkg.Options{}
			if name == "resource" || name == "data" {
				opts.Schemas = schemas
			}
			if err := alignpkg.Apply(file, opts); err != nil {
				t.Fatalf("align fmt: %v", err)
			}
			gotAligned := file.Bytes()
			if !bytes.Equal(gotAligned, alignedBytes) {
				t.Fatalf("aligned mismatch for %s:\n-- got --\n%s\n-- want --\n%s", name, gotAligned, alignedBytes)
			}

			file2, diags := hclwrite.ParseConfig(alignedBytes, alignedPath, hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse aligned: %v", diags)
			}
			if err := alignpkg.Apply(file2, opts); err != nil {
				t.Fatalf("align aligned: %v", err)
			}
			if !bytes.Equal(alignedBytes, file2.Bytes()) {
				t.Fatalf("non-idempotent on aligned for %s", name)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk cases: %v", err)
	}
}

func TestUnknownAttributesOrder(t *testing.T) {
	src := []byte(`variable "example" {
  foo = "foo"
  description = "example"
  bar = "bar"
  type = number
  default = 1
}`)

	t.Run("original", func(t *testing.T) {
		file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse input: %v", diags)
		}
		if err := alignpkg.Apply(file, &alignpkg.Options{}); err != nil {
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
	})

	t.Run("prefix", func(t *testing.T) {
		file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse input: %v", diags)
		}
		if err := alignpkg.Apply(file, &alignpkg.Options{PrefixOrder: true}); err != nil {
			t.Fatalf("reorder: %v", err)
		}
		got := file.Bytes()
		exp := []byte(`variable "example" {
  description = "example"
  type        = number
  default     = 1
  bar         = "bar"
  foo         = "foo"
}`)
		if !bytes.Equal(got, exp) {
			t.Fatalf("output mismatch:\n-- got --\n%s\n-- want --\n%s", got, exp)
		}
	})
}
