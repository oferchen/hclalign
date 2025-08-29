// internal/align/locals.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type localsStrategy struct{}

func (localsStrategy) Name() string { return "locals" }

func (localsStrategy) Align(block *hclwrite.Block, _ *Options) error {
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	return reorderBlock(block, names)
}

func init() { Register(localsStrategy{}) }
