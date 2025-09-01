// internal/align/schema/loader_test.go
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
              "r1": {"required": true},
              "r2": {"required": true},
              "o1": {"optional": true},
              "o2": {"optional": true},
              "c1": {"computed": true},
              "c2": {"computed": true}
            },
            "block_types": {
              "first": {
                "block": {
                  "attributes": {"x": {"required": true}}
                }
              },
              "second": {
                "block": {
                  "attributes": {"y": {"optional": true}}
                }
              }
            }
          }
        }
      }
    },
    "registry.terraform.io/acme/test": {
      "resource_schemas": {
        "test_thing": {
          "block": {
            "attributes": {
              "foo": {"required": true}
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

	k1 := "registry.terraform.io/hashicorp/test/test_thing"
	s, ok := schemas[k1]
	require.True(t, ok)
	require.Equal(t, []string{"r1", "r2"}, s.RequiredOrder)
	require.Equal(t, []string{"o1", "o2"}, s.OptionalOrder)
	require.Equal(t, []string{"c1", "c2"}, s.ComputedOrder)
	require.Equal(t, []string{"first", "second"}, s.BlocksOrder)
	first, ok := s.Blocks["first"]
	require.True(t, ok)
	require.Equal(t, []string{"x"}, first.RequiredOrder)
	second, ok := s.Blocks["second"]
	require.True(t, ok)
	require.Equal(t, []string{"y"}, second.OptionalOrder)

	k2 := "registry.terraform.io/acme/test/test_thing"
	_, ok = schemas[k2]
	require.True(t, ok)
}

func TestFromTerraformCaching(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	samplePath := filepath.Join(dir, "sample.json")
	versionPath := filepath.Join(dir, "version.json")
	require.NoError(t, os.WriteFile(samplePath, []byte(sample), 0o644))
	version := `{"terraform_version":"1.0.0","provider_selections":{"registry.terraform.io/hashicorp/test":{"version":"1.0.0"}}}`
	require.NoError(t, os.WriteFile(versionPath, []byte(version), 0o644))

	var providersCalls int
	orig := execCommandContext
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if len(args) > 0 && args[0] == "version" {
			return exec.CommandContext(ctx, "cat", versionPath)
		}
		providersCalls++
		return exec.CommandContext(ctx, "cat", samplePath)
	}
	defer func() { execCommandContext = orig }()

	ctx := context.Background()
	_, err := FromTerraform(ctx, cacheDir, dir, false)
	require.NoError(t, err)
	_, err = FromTerraform(ctx, cacheDir, dir, false)
	require.NoError(t, err)
	require.Equal(t, 1, providersCalls)

	version = `{"terraform_version":"1.0.0","provider_selections":{"registry.terraform.io/hashicorp/test":{"version":"2.0.0"}}}`
	require.NoError(t, os.WriteFile(versionPath, []byte(version), 0o644))
	_, err = FromTerraform(ctx, cacheDir, dir, false)
	require.NoError(t, err)
	require.Equal(t, 2, providersCalls)
}

func TestFromTerraformNoCache(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	samplePath := filepath.Join(dir, "sample.json")
	versionPath := filepath.Join(dir, "version.json")
	require.NoError(t, os.WriteFile(samplePath, []byte(sample), 0o644))
	version := `{"terraform_version":"1.0.0","provider_selections":{}}`
	require.NoError(t, os.WriteFile(versionPath, []byte(version), 0o644))

	var providersCalls int
	orig := execCommandContext
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if len(args) > 0 && args[0] == "version" {
			return exec.CommandContext(ctx, "cat", versionPath)
		}
		providersCalls++
		return exec.CommandContext(ctx, "cat", samplePath)
	}
	defer func() { execCommandContext = orig }()

	ctx := context.Background()
	_, err := FromTerraform(ctx, cacheDir, dir, true)
	require.NoError(t, err)
	_, err = FromTerraform(ctx, cacheDir, dir, true)
	require.NoError(t, err)
	require.Equal(t, 2, providersCalls)
}
