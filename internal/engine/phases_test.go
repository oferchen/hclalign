// internal/engine/phases_test.go
package engine

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestPhases(t *testing.T) {
	cases := []string{"simple", "heredocs", "templates", "trailing_commas", "comments", "locals", "output", "module", "provider", "terraform", "resource", "data", "idempotency"}
	base := filepath.Join("..", "..", "tests", "cases")
	schemaPath := filepath.Join("..", "..", "tests", "testdata", "providers-schema.json")

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			inBytes, err := os.ReadFile(filepath.Join(base, name, "in.tf"))
			require.NoError(t, err)
			outExp, err := os.ReadFile(filepath.Join(base, name, "out.tf"))
			require.NoError(t, err)

			cfg := &config.Config{Stdout: true}
			if name == "resource" || name == "data" {
				cfg.ProvidersSchema = schemaPath
			}

			var out bytes.Buffer
			changed, err := ProcessReader(context.Background(), bytes.NewReader(inBytes), &out, cfg)
			require.NoError(t, err)
			require.Equal(t, !bytes.Equal(inBytes, outExp), changed)
			require.Equal(t, string(outExp), out.String())

			out.Reset()
			changed, err = ProcessReader(context.Background(), bytes.NewReader(outExp), &out, cfg)
			require.NoError(t, err)
			require.False(t, changed)
			require.Equal(t, string(outExp), out.String())
		})
	}

	t.Run("error", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{Stdout: true}
		_, err := ProcessReader(context.Background(), bytes.NewReader([]byte("variable \"a\" {")), &out, cfg)
		require.Error(t, err)
	})
}
