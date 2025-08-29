// /internal/align/schema/loader_test.go
package schema

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const sample = `{
  "provider_schemas": {
    "registry.terraform.io/hashicorp/test": {
      "resource_schemas": {
        "test_thing": {
          "block": {
            "attributes": {
              "foo": {"required": true},
              "bar": {"optional": true},
              "baz": {"computed": true}
            }
          }
        }
      },
      "data_source_schemas": {
        "test_data": {
          "block": {
            "attributes": {
              "id": {"computed": true}
            }
          }
        }
      }
    }
  }
}`

func TestLoad(t *testing.T) {
	r := strings.NewReader(sample)
	schemas, err := Load(r)
	require.NoError(t, err)

	s, ok := schemas["test_thing"]
	require.True(t, ok)
	_, req := s.Required["foo"]
	require.True(t, req)
	_, opt := s.Optional["bar"]
	require.True(t, opt)
	_, comp := s.Computed["baz"]
	require.True(t, comp)
	_, meta := s.Meta["provider"]
	require.True(t, meta)

	ds, ok := schemas["test_data"]
	require.True(t, ok)
	_, comp2 := ds.Computed["id"]
	require.True(t, comp2)
}

func TestFromTerraformCaching(t *testing.T) {
	dir := t.TempDir()
	cache := filepath.Join(dir, "schema.json")
	samplePath := filepath.Join(dir, "sample.json")
	require.NoError(t, os.WriteFile(samplePath, []byte(sample), 0o644))

	var calls int
	orig := execCommandContext
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls++
		return exec.CommandContext(ctx, "cat", samplePath)
	}
	defer func() { execCommandContext = orig }()

	ctx := context.Background()
	_, err := FromTerraform(ctx, cache)
	require.NoError(t, err)
	_, err = FromTerraform(ctx, cache)
	require.NoError(t, err)
	require.Equal(t, 1, calls)
}
