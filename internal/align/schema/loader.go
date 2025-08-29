// /internal/align/schema/loader.go
package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/oferchen/hclalign/internal/align"
)

func Load(r io.Reader) (map[string]*align.Schema, error) {
	var root tfSchema
	dec := json.NewDecoder(r)
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}
	result := make(map[string]*align.Schema)
	for _, ps := range root.ProviderSchemas {
		for typ, s := range ps.ResourceSchemas {
			result[typ] = buildSchema(s.Block.Attributes)
		}
		for typ, s := range ps.DataSourceSchemas {
			result[typ] = buildSchema(s.Block.Attributes)
		}
	}
	return result, nil
}

type tfSchema struct {
	ProviderSchemas map[string]providerSchema `json:"provider_schemas"`
}

type providerSchema struct {
	ResourceSchemas   map[string]schemaBlock `json:"resource_schemas"`
	DataSourceSchemas map[string]schemaBlock `json:"data_source_schemas"`
}

type schemaBlock struct {
	Block block `json:"block"`
}

type block struct {
	Attributes map[string]attribute `json:"attributes"`
}

type attribute struct {
	Required bool `json:"required"`
	Optional bool `json:"optional"`
	Computed bool `json:"computed"`
}

func buildSchema(attrs map[string]attribute) *align.Schema {
	s := &align.Schema{
		Required: map[string]struct{}{},
		Optional: map[string]struct{}{},
		Computed: map[string]struct{}{},
		Meta:     map[string]struct{}{"count": {}, "for_each": {}, "provider": {}, "depends_on": {}},
	}
	for name, attr := range attrs {
		switch {
		case attr.Required:
			s.Required[name] = struct{}{}
		case attr.Optional:
			s.Optional[name] = struct{}{}
		case attr.Computed:
			s.Computed[name] = struct{}{}
		}
	}
	return s
}

func LoadFile(path string) (map[string]*align.Schema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Load(f)
}

var execCommandContext = exec.CommandContext

func FromTerraform(ctx context.Context, cachePath string) (map[string]*align.Schema, error) {
	if b, err := os.ReadFile(cachePath); err == nil {
		return Load(bytes.NewReader(b))
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	cmd := execCommandContext(ctx, "terraform", "providers", "schema", "-json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("terraform providers schema: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(cachePath, out, 0o644); err != nil {
		return nil, err
	}
	return Load(bytes.NewReader(out))
}
