// internal/align/provider.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type providerStrategy struct{}

func (providerStrategy) Name() string { return "provider" }

func (providerStrategy) Align(block *hclwrite.Block, _ *Options) error {
	attrs := block.Body().Attributes()
	canonical := CanonicalBlockAttrOrder["provider"]

	names := make([]string, 0, len(attrs))
	reserved := make(map[string]struct{}, len(canonical))
	for _, name := range canonical {
		if _, ok := attrs[name]; ok {
			names = append(names, name)
		}
		reserved[name] = struct{}{}
	}

	original := ihcl.AttributeOrder(block.Body(), attrs)
	for _, name := range original {
		if _, ok := reserved[name]; ok {
			continue
		}
		names = append(names, name)
	}
	return reorderBlock(block, names)
}

func init() { Register(providerStrategy{}) }
