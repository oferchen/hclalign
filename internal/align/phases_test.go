// internal/align/phases_test.go
package align_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/hashicorp/hclalign/internal/align"
	alignschema "github.com/hashicorp/hclalign/internal/align/schema"
	terraformfmt "github.com/hashicorp/hclalign/internal/fmt"
	internalfs "github.com/hashicorp/hclalign/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestPhases(t *testing.T) {
	cases := []string{"simple", "heredocs", "templates", "trailing_commas", "locals", "output", "module", "provider", "terraform", "resource", "data"}
	base := filepath.Join("..", "..", "tests", "cases")
	schemaPath := filepath.Join("..", "..", "tests", "testdata", "providers-schema.json")
	schemas, err := alignschema.LoadFile(schemaPath)
	require.NoError(t, err)

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			inPath := filepath.Join(base, name, "in.tf")
			fmtPath := filepath.Join(base, name, "fmt.tf")
			alignedPath := filepath.Join(base, name, "aligned.tf")

			inBytes, err := os.ReadFile(inPath)
			require.NoError(t, err)
			wantFmt, err := os.ReadFile(fmtPath)
			require.NoError(t, err)
			wantAligned, err := os.ReadFile(alignedPath)
			require.NoError(t, err)

			gotFmt, err := terraformfmt.Format(inBytes, inPath, string(terraformfmt.StrategyGo))
			require.NoError(t, err)
			hadNewline := len(inBytes) > 0 && inBytes[len(inBytes)-1] == '\n'
			if !hadNewline && len(gotFmt) > 0 && gotFmt[len(gotFmt)-1] == '\n' {
				gotFmt = gotFmt[:len(gotFmt)-1]
			}
			require.Equal(t, string(wantFmt), string(gotFmt))

			hints := internalfs.DetectHintsFromBytes(gotFmt)
			parseData := internalfs.PrepareForParse(gotFmt, hints)
			file, diags := hclwrite.ParseConfig(parseData, inPath, hcl.InitialPos)
			require.False(t, diags.HasErrors())

			opts := &alignpkg.Options{}
			if name == "resource" || name == "data" {
				opts.Schemas = schemas
			}
			require.NoError(t, alignpkg.Apply(file, opts))
			gotAligned := hclwrite.Format(file.Bytes())
			require.Equal(t, string(wantAligned), string(gotAligned))

			parseData2 := internalfs.PrepareForParse(gotAligned, internalfs.DetectHintsFromBytes(gotAligned))
			file2, diags := hclwrite.ParseConfig(parseData2, alignedPath, hcl.InitialPos)
			require.False(t, diags.HasErrors())
			require.NoError(t, alignpkg.Apply(file2, opts))
			require.Equal(t, string(gotAligned), string(hclwrite.Format(file2.Bytes())))
		})
	}

	t.Run("error", func(t *testing.T) {
		_, err := terraformfmt.Format([]byte("variable \"a\" {"), "bad.hcl", string(terraformfmt.StrategyGo))
		require.Error(t, err)
	})
}
