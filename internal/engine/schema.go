package engine

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/hclalign/config"
	"github.com/hashicorp/hclalign/internal/align"
	alignschema "github.com/hashicorp/hclalign/internal/align/schema"
)

// loadSchemas loads provider schemas based on configuration. It returns nil if
// no schema options are provided.
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
