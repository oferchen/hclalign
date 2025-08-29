// internal/align/output.go
package align

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type outputStrategy struct{}

func (outputStrategy) Name() string { return "output" }

var outputCanonicalOrder = []string{"description", "value", "sensitive", "depends_on"}

func (outputStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()

	order := make([]string, 0, len(attrs))
	reserved := make(map[string]struct{}, len(outputCanonicalOrder))
	for _, name := range outputCanonicalOrder {
		if _, ok := attrs[name]; ok {
			order = append(order, name)
		}
		reserved[name] = struct{}{}
	}

	original := ihcl.AttributeOrder(block.Body(), attrs)

	if opts != nil && opts.Strict {
		var unknown []string
		for _, name := range original {
			if _, ok := reserved[name]; !ok {
				unknown = append(unknown, name)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			return fmt.Errorf("output: unknown attributes: %s", strings.Join(unknown, ", "))
		}
	}

	for _, name := range original {
		if _, ok := reserved[name]; !ok {
			order = append(order, name)
		}
	}

	return reorderBlock(block, order)
}

func init() { Register(outputStrategy{}) }
