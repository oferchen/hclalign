package align

import "github.com/hashicorp/hcl/v2/hclwrite"

type dataStrategy struct{}

func (dataStrategy) Name() string { return "data" }

func (dataStrategy) Align(block *hclwrite.Block, opts *Options) error {
	return schemaAwareOrder(block, opts)
}

func init() { Register(dataStrategy{}) }
