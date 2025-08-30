// internal/align/phases_test.go
package align_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	alignschema "github.com/oferchen/hclalign/internal/align/schema"
	terraformfmt "github.com/oferchen/hclalign/internal/fmt"
	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestPhases(t *testing.T) {
	cases := []string{"simple", "heredocs", "templates", "trailing_commas", "comments", "locals", "output", "module", "provider", "terraform", "resource", "data", "idempotency"}
	base := filepath.Join("..", "..", "tests", "cases")
	schemaPath := filepath.Join("..", "..", "tests", "testdata", "providers-schema.json")
	schemas, err := alignschema.LoadFile(schemaPath)
	require.NoError(t, err)

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			inPath := filepath.Join(base, name, "in.tf")
			outPath := filepath.Join(base, name, "out.tf")

			inBytes, err := os.ReadFile(inPath)
			require.NoError(t, err)
			wantOut, err := os.ReadFile(outPath)
			require.NoError(t, err)

			fmtBytes, _, err := terraformfmt.Format(inBytes, inPath, string(terraformfmt.StrategyGo))
			require.NoError(t, err)

                        fmtBytes, _, err := terraformfmt.Format(inBytes, inPath, string(terraformfmt.StrategyGo))
                        require.NoError(t, err)

			hadNewline := len(inBytes) > 0 && inBytes[len(inBytes)-1] == '\n'
			if !hadNewline && len(fmtBytes) > 0 && fmtBytes[len(fmtBytes)-1] == '\n' {
				fmtBytes = fmtBytes[:len(fmtBytes)-1]
			}
			againFmt, _, err := terraformfmt.Format(fmtBytes, inPath, string(terraformfmt.StrategyGo))
			require.NoError(t, err)

                        againFmt, _, err := terraformfmt.Format(fmtBytes, inPath, string(terraformfmt.StrategyGo))
                        require.NoError(t, err)
			hadFmtNewline := len(fmtBytes) > 0 && fmtBytes[len(fmtBytes)-1] == '\n'
			if !hadFmtNewline && len(againFmt) > 0 && againFmt[len(againFmt)-1] == '\n' {
				againFmt = againFmt[:len(againFmt)-1]
			}
			require.Equal(t, string(fmtBytes), string(againFmt))

			hints := internalfs.DetectHintsFromBytes(fmtBytes)
			parseData := internalfs.PrepareForParse(fmtBytes, hints)
			file, diags := hclwrite.ParseConfig(parseData, inPath, hcl.InitialPos)
			require.False(t, diags.HasErrors())

			opts := &alignpkg.Options{}
			if name == "resource" || name == "data" {
				opts.Schemas = schemas
			}
			require.NoError(t, alignpkg.Apply(file, opts))
			gotAligned := hclwrite.Format(file.Bytes())
			require.Equal(t, string(wantOut), string(gotAligned))

			parseData2 := internalfs.PrepareForParse(gotAligned, internalfs.DetectHintsFromBytes(gotAligned))
			file2, diags := hclwrite.ParseConfig(parseData2, outPath, hcl.InitialPos)
			require.False(t, diags.HasErrors())
			require.NoError(t, alignpkg.Apply(file2, opts))
			require.Equal(t, string(gotAligned), string(hclwrite.Format(file2.Bytes())))
		})
	}

	t.Run("error", func(t *testing.T) {
		_, _, err := terraformfmt.Format([]byte("variable \"a\" {"), "bad.hcl", string(terraformfmt.StrategyGo))
		require.Error(t, err)
	})

                _, _, err := terraformfmt.Format([]byte("variable \"a\" {"), "bad.hcl", string(terraformfmt.StrategyGo))
                require.Error(t, err)
        })
}
