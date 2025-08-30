// internal/align/dynamic.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type dynamicStrategy struct{}

func (dynamicStrategy) Name() string { return "dynamic" }

func (dynamicStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	allowed := map[string]struct{}{"for_each": {}, "iterator": {}}
	var known, unknown []string
	for name := range attrs {
		if _, ok := allowed[name]; ok {
			known = append(known, name)
		} else {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(known)
	sort.Strings(unknown)
	names := append(known, unknown...)
	return reorderBlock(block, names)
}

func init() { Register(dynamicStrategy{}) }
