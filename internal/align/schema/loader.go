// internal/align/schema/loader.go
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
	for prov, ps := range root.ProviderSchemas {
		for typ, s := range ps.ResourceSchemas {
			key := prov + "/" + typ
			result[key] = buildSchema(s.Block)
		}
		for typ, s := range ps.DataSourceSchemas {
			key := prov + "/" + typ
			result[key] = buildSchema(s.Block)
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
	Attributes orderedAttributes `json:"attributes"`
	BlockTypes orderedBlockTypes `json:"block_types"`
}

type orderedAttributes struct {
	Order []string
	Items map[string]attribute
}

func (o *orderedAttributes) UnmarshalJSON(data []byte) error {
	o.Items = map[string]attribute{}
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil || t != json.Delim('{') {
		return err
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		key := tok.(string)
		o.Order = append(o.Order, key)
		var a attribute
		if err := dec.Decode(&a); err != nil {
			return err
		}
		o.Items[key] = a
	}
	_, err = dec.Token()
	return err
}

type orderedBlockTypes struct {
	Order []string
	Items map[string]schemaBlock
}

func (o *orderedBlockTypes) UnmarshalJSON(data []byte) error {
	o.Items = map[string]schemaBlock{}
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil || t != json.Delim('{') {
		return err
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		key := tok.(string)
		o.Order = append(o.Order, key)
		var sb schemaBlock
		if err := dec.Decode(&sb); err != nil {
			return err
		}
		o.Items[key] = sb
	}
	_, err = dec.Token()
	return err
}

type attribute struct {
	Required bool `json:"required"`
	Optional bool `json:"optional"`
	Computed bool `json:"computed"`
}

func buildSchema(b block) *align.Schema {
	s := &align.Schema{
		Required: map[string]struct{}{},
		Optional: map[string]struct{}{},
		Computed: map[string]struct{}{},
		Meta:     map[string]struct{}{"count": {}, "for_each": {}, "provider": {}, "depends_on": {}},
		Blocks:   map[string]*align.Schema{},
	}
	for _, name := range b.Attributes.Order {
		attr := b.Attributes.Items[name]
		switch {
		case attr.Required:
			s.Required[name] = struct{}{}
			s.RequiredOrder = append(s.RequiredOrder, name)
		case attr.Optional:
			s.Optional[name] = struct{}{}
			s.OptionalOrder = append(s.OptionalOrder, name)
		case attr.Computed:
			s.Computed[name] = struct{}{}
			s.ComputedOrder = append(s.ComputedOrder, name)
		}
	}
	for _, name := range b.BlockTypes.Order {
		sb := b.BlockTypes.Items[name]
		s.BlocksOrder = append(s.BlocksOrder, name)
		s.Blocks[name] = buildSchema(sb.Block)
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
