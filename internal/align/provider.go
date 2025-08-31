// internal/align/provider.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
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

	others := make([]string, 0, len(attrs))
	for name := range attrs {
		if _, ok := reserved[name]; ok {
			continue
		}
		others = append(others, name)
	}
	sort.Strings(others)
	names = append(names, others...)

	return reorderBlock(block, names)
}

func init() { Register(providerStrategy{}) }
