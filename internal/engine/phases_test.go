package engine

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hclalign/config"
	terraformfmt "github.com/hashicorp/hclalign/internal/fmt"
	"github.com/stretchr/testify/require"
)

func TestPhases(t *testing.T) {
	cases := []string{"simple", "heredocs", "templates", "trailing_commas", "locals", "output", "module", "provider", "terraform", "resource", "data"}
	base := filepath.Join("..", "..", "tests", "cases")
	schemaPath := filepath.Join("..", "..", "tests", "testdata", "providers-schema.json")

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			inBytes, err := os.ReadFile(filepath.Join(base, name, "in.tf"))
			require.NoError(t, err)
			fmtExp, err := os.ReadFile(filepath.Join(base, name, "fmt.tf"))
			require.NoError(t, err)
			alignedExp, err := os.ReadFile(filepath.Join(base, name, "aligned.tf"))
			require.NoError(t, err)

			cfgFmt := &config.Config{Stdout: true, FmtOnly: true, FmtStrategy: string(terraformfmt.StrategyGo)}
			cfgFull := &config.Config{Stdout: true, FmtStrategy: string(terraformfmt.StrategyGo)}
			if name == "resource" || name == "data" {
				cfgFmt.ProvidersSchema = schemaPath
				cfgFull.ProvidersSchema = schemaPath
			}

			var out bytes.Buffer
			changed, err := ProcessReader(context.Background(), bytes.NewReader(inBytes), &out, cfgFmt)
			require.NoError(t, err)
			require.Equal(t, !bytes.Equal(inBytes, fmtExp), changed)
			require.Equal(t, string(fmtExp), out.String())

			out.Reset()
			changed, err = ProcessReader(context.Background(), bytes.NewReader(inBytes), &out, cfgFull)
			require.NoError(t, err)
			require.Equal(t, !bytes.Equal(inBytes, alignedExp), changed)
			require.Equal(t, string(alignedExp), out.String())

			out.Reset()
			changed, err = ProcessReader(context.Background(), bytes.NewReader(alignedExp), &out, cfgFull)
			require.NoError(t, err)
			require.False(t, changed)
			require.Equal(t, string(alignedExp), out.String())
		})
	}

	t.Run("error", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{Stdout: true, FmtStrategy: string(terraformfmt.StrategyGo)}
		_, err := ProcessReader(context.Background(), bytes.NewReader([]byte("variable \"a\" {")), &out, cfg)
		require.Error(t, err)
	})
}
