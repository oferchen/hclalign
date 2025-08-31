// internal/align/golden_test.go
package align_test

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	alignschema "github.com/oferchen/hclalign/internal/align/schema"
	terraformfmt "github.com/oferchen/hclalign/internal/fmt"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

func TestGolden(t *testing.T) {
	casesDir := filepath.Join("..", "..", "tests", "cases")
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
		if path == casesDir {
			return nil
		}
		name, err := filepath.Rel(casesDir, path)
		if err != nil {
			return err
		}

		t.Run(name, func(t *testing.T) {
			inPath := filepath.Join(path, "in.tf")
			outPath := filepath.Join(path, "out.tf")
			if _, err := os.Stat(inPath); err != nil {
				t.Skip("missing in.tf")
			}
			if _, err := os.Stat(outPath); err != nil {
				t.Skip("missing out.tf")
			}

			inBytes, err := os.ReadFile(inPath)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}
			outBytes, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("read out: %v", err)
			}

			fmtBytes, _, err := terraformfmt.Format(context.Background(), inBytes, inPath, string(terraformfmt.StrategyGo))
			if err != nil {
				t.Fatalf("format input: %v", err)
			}

			fmtBytes, _, err := terraformfmt.Format(context.Background(), inBytes, inPath, string(terraformfmt.StrategyGo))
			if err != nil {
				t.Fatalf("format input: %v", err)
			}
			hadNewline := len(inBytes) > 0 && inBytes[len(inBytes)-1] == '\n'
			if !hadNewline && len(fmtBytes) > 0 && fmtBytes[len(fmtBytes)-1] == '\n' {
				fmtBytes = fmtBytes[:len(fmtBytes)-1]
			}

			againFmt, _, err := terraformfmt.Format(context.Background(), fmtBytes, inPath, string(terraformfmt.StrategyGo))
			if err != nil {
				t.Fatalf("format fmt: %v", err)
			}

			againFmt, _, err := terraformfmt.Format(context.Background(), fmtBytes, inPath, string(terraformfmt.StrategyGo))
			if err != nil {
				t.Fatalf("format fmt: %v", err)
			}
			hadFmtNewline := len(fmtBytes) > 0 && fmtBytes[len(fmtBytes)-1] == '\n'
			if !hadFmtNewline && len(againFmt) > 0 && againFmt[len(againFmt)-1] == '\n' {
				againFmt = againFmt[:len(againFmt)-1]
			}
			if !bytes.Equal(againFmt, fmtBytes) {
				t.Fatalf("fmt not idempotent for %s", name)
			}

			hints := internalfs.DetectHintsFromBytes(fmtBytes)
			parseBytes := internalfs.PrepareForParse(fmtBytes, hints)
			file, diags := hclwrite.ParseConfig(parseBytes, inPath, hcl.InitialPos)
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
			if !bytes.Equal(gotAligned, outBytes) {
				t.Fatalf("aligned mismatch for %s:\n-- got --\n%s\n-- want --\n%s", name, gotAligned, outBytes)
			}

			file2, diags := hclwrite.ParseConfig(outBytes, outPath, hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parse out: %v", diags)
			}
			if err := alignpkg.Apply(file2, opts); err != nil {
				t.Fatalf("align out: %v", err)
			}
			if !bytes.Equal(outBytes, file2.Bytes()) {
				t.Fatalf("non-idempotent on out for %s", name)
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
