package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type providerStrategy struct{}

func (providerStrategy) Name() string { return "provider" }

func (providerStrategy) Align(block *hclwrite.Block, _ *Options) error {
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	return reorderBlock(block, names)
}

func init() { Register(providerStrategy{}) }
