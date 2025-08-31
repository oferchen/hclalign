// internal/engine/schema_test.go
package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/config"
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

func TestLoadSchemasNil(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	schemas, err := loadSchemas(context.Background(), cfg)
	require.NoError(t, err)
	require.Nil(t, schemas)
}

func TestLoadSchemasProvidersSchema(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "..", "tests", "testdata", "providers-schema.json")
	cfg := &config.Config{ProvidersSchema: path}
	schemas, err := loadSchemas(context.Background(), cfg)
	require.NoError(t, err)
	_, ok := schemas["aws_s3_bucket"]
	require.True(t, ok)
}

func TestLoadSchemasUseTerraformSchema(t *testing.T) {
	dir := t.TempDir()
	samplePath := filepath.Join(dir, "sample.json")
	require.NoError(t, os.WriteFile(samplePath, []byte(sample), 0o644))
	script := filepath.Join(dir, "terraform")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\n/bin/cat "+samplePath+"\n"), 0o755))
	t.Setenv("PATH", dir)
	cfg := &config.Config{UseTerraformSchema: true, ProvidersSchema: filepath.Join(dir, "schema.json")}
	schemas, err := loadSchemas(context.Background(), cfg)
	require.NoError(t, err)
	_, ok := schemas["test_thing"]
	require.True(t, ok)
	_, ok = schemas["test_data"]
	require.True(t, ok)
}

func TestLoadSchemasUseTerraformSchemaError(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "terraform")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nexit 1\n"), 0o755))
	t.Setenv("PATH", dir)
	cfg := &config.Config{UseTerraformSchema: true, ProvidersSchema: filepath.Join(dir, "schema.json")}
	schemas, err := loadSchemas(context.Background(), cfg)
	require.Error(t, err)
	require.Nil(t, schemas)
}

func TestLoadSchemasUseTerraformSchemaCache(t *testing.T) {
	dir := t.TempDir()
	samplePath := filepath.Join(dir, "sample.json")
	require.NoError(t, os.WriteFile(samplePath, []byte(sample), 0o644))
	script := filepath.Join(dir, "terraform")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\n/bin/cat "+samplePath+"\n"), 0o755))
	cache := filepath.Join(dir, "schema.json")
	t.Setenv("PATH", dir)
	cfg := &config.Config{UseTerraformSchema: true, ProvidersSchema: cache}
	_, err := loadSchemas(context.Background(), cfg)
	require.NoError(t, err)
	require.NoError(t, os.Remove(script))
	t.Setenv("PATH", "")
	schemas, err := loadSchemas(context.Background(), cfg)
	require.NoError(t, err)
	_, ok := schemas["test_thing"]
	require.True(t, ok)
}

func TestLoadSchemasMissingFile(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{ProvidersSchema: filepath.Join(t.TempDir(), "missing.json")}
	schemas, err := loadSchemas(context.Background(), cfg)
	require.Error(t, err)
	require.Nil(t, schemas)
}
