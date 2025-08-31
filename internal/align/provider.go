// internal/align/provider.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type providerStrategy struct{}

func (providerStrategy) Name() string { return "provider" }

func (providerStrategy) Align(block *hclwrite.Block, opts *Options) error {
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
	extra := make([]string, 0)
	for _, name := range original {
		if _, ok := reserved[name]; ok {
			continue
		}
		extra = append(extra, name)
	}
	names = append(names, extra...)
	return reorderBlock(block, names)
}

func init() { Register(providerStrategy{}) }
