// /internal/align/locals.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type localsStrategy struct{}

func (localsStrategy) Name() string { return "locals" }

func (localsStrategy) Align(block *hclwrite.Block, opts *Options) error {
	if opts == nil || opts.BlockOrder == nil || opts.BlockOrder["locals"] != "alphabetical" {
		return nil
	}

	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	for n := range attrs {
		names = append(names, n)
	}
	sort.Strings(names)
	return reorderBlock(block, names)
}

func init() { Register(localsStrategy{}) }
