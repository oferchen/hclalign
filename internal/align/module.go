// filename: internal/align/module.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type moduleStrategy struct{}

func (moduleStrategy) Name() string { return "module" }

func (moduleStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()

	canonical := CanonicalBlockAttrOrder["module"]
	order := make([]string, 0, len(attrs))
	reserved := make(map[string]struct{}, len(canonical))
	for _, name := range canonical {
		if _, ok := attrs[name]; ok {
			order = append(order, name)
		}
		reserved[name] = struct{}{}
	}

	vars := make([]string, 0, len(attrs))
	for name := range attrs {
		if _, ok := reserved[name]; !ok {
			vars = append(vars, name)
		}
	}
	sort.Strings(vars)

	order = append(order, vars...)

	return reorderBlock(block, order)
}

func init() { Register(moduleStrategy{}) }
