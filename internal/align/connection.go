// internal/align/connection.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type connectionStrategy struct{}

func (connectionStrategy) Name() string { return "connection" }

func (connectionStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	return reorderBlock(block, names)
}

func init() { Register(connectionStrategy{}) }
