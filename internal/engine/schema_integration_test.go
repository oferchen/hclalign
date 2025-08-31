// internal/engine/schema_integration_test.go
package engine

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestProcessReaderWithSchema(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.json")
	require.NoError(t, os.WriteFile(schemaPath, []byte(sample), 0o644))

	cfg := &config.Config{ProvidersSchema: schemaPath, Stdout: true}
	input := `resource "test_thing" "x" {
  baz = 1
  foo = 2
  bar = 3
}
`
	var out bytes.Buffer
	changed, err := processReader(context.Background(), strings.NewReader(input), &out, cfg)
	require.NoError(t, err)
	require.True(t, changed)
	want := "resource \"test_thing\" \"x\" {\n  foo = 2\n  bar = 3\n  baz = 1\n}\n"
	require.Equal(t, want, out.String())
}
