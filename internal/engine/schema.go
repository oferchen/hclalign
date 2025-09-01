// internal/engine/schema.go
package engine

import (
	"context"
	"os"
	"path/filepath"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/align"
	alignschema "github.com/oferchen/hclalign/internal/align/schema"
)

func loadSchemas(ctx context.Context, cfg *config.Config) (map[string]*align.Schema, error) {
	if cfg.ProvidersSchema == "" && !cfg.UseTerraformSchema {
		return nil, nil
	}
	if cfg.ProvidersSchema != "" {
		return alignschema.LoadFile(cfg.ProvidersSchema)
	}
	modulePath := cfg.Target
	if modulePath == "" {
		modulePath = "."
	}
	if fi, err := os.Stat(modulePath); err == nil && !fi.IsDir() {
		modulePath = filepath.Dir(modulePath)
	}
	cacheDir := cfg.SchemaCache
	if cacheDir == "" {
		cacheDir = filepath.Join(modulePath, ".terraform", "schema-cache")
	}
	return alignschema.FromTerraform(ctx, cacheDir, modulePath, cfg.NoSchemaCache)
}
