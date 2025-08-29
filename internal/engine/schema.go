// /internal/engine/schema.go
package engine

import (
	"context"
	"path/filepath"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/align"
	alignschema "github.com/oferchen/hclalign/internal/align/schema"
)

func loadSchemas(ctx context.Context, cfg *config.Config) (map[string]*align.Schema, error) {
	if cfg.ProvidersSchema == "" && !cfg.UseTerraformSchema {
		return nil, nil
	}
	path := cfg.ProvidersSchema
	if path == "" {
		path = filepath.Join(".terraform", "providers-schema.json")
	}
	if cfg.UseTerraformSchema {
		return alignschema.FromTerraform(ctx, path)
	}
	return alignschema.LoadFile(path)
}
