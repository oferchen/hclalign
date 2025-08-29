package schema

import (
	"encoding/json"
	"io"
)

type ProvidersSchema struct {
	ProviderSchemas map[string]*ProviderSchema `json:"provider_schemas"`
}

type ProviderSchema struct {
	ResourceSchemas   map[string]*ResourceSchema `json:"resource_schemas"`
	DataSourceSchemas map[string]*ResourceSchema `json:"data_source_schemas"`
}

type ResourceSchema struct {
	Block *Block `json:"block"`
}

type Block struct {
	Attributes map[string]*Attribute `json:"attributes"`
}

type Attribute struct {
	Required bool `json:"required"`
	Optional bool `json:"optional"`
	Computed bool `json:"computed"`
}

func Parse(r io.Reader) (*ProvidersSchema, error) {
	var ps ProvidersSchema
	dec := json.NewDecoder(r)
	if err := dec.Decode(&ps); err != nil {
		return nil, err
	}
	return &ps, nil
}
