// internal/align/module.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type moduleStrategy struct{}

func (moduleStrategy) Name() string { return "module" }

func (moduleStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()

	order := make([]string, 0, len(attrs))

	if _, ok := attrs["source"]; ok {
		order = append(order, "source")
	}
	if _, ok := attrs["version"]; ok {
		order = append(order, "version")
	}

	metaArgs := []string{"providers", "count", "for_each", "depends_on"}
	for _, name := range metaArgs {
		if _, ok := attrs[name]; ok {
			order = append(order, name)
		}
	}

	reserved := map[string]struct{}{
		"source":     {},
		"version":    {},
		"providers":  {},
		"count":      {},
		"for_each":   {},
		"depends_on": {},
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
