// internal/align/output.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type outputStrategy struct{}

func (outputStrategy) Name() string { return "output" }

func (outputStrategy) Align(block *hclwrite.Block, opts *Options) error {
	body := block.Body()
	attrs := body.Attributes()

	canonical := CanonicalBlockAttrOrder["output"]
	order := make([]string, 0, len(attrs))
	reserved := make(map[string]struct{}, len(canonical))
	for _, name := range canonical {
		if _, ok := attrs[name]; ok {
			order = append(order, name)
		}
		reserved[name] = struct{}{}
	}

	original := ihcl.AttributeOrder(body, attrs)

	nonCanonical := make([]string, 0)
	for _, name := range original {
		if _, ok := reserved[name]; !ok {
			nonCanonical = append(nonCanonical, name)
		}
	}
	order = append(order, nonCanonical...)

	nested := body.Blocks()
	if len(nested) > 0 {
		pre := make([]*hclwrite.Block, 0)
		post := make([]*hclwrite.Block, 0)
		other := make([]*hclwrite.Block, 0)
		for _, nb := range nested {
			switch nb.Type() {
			case "precondition":
				pre = append(pre, nb)
			case "postcondition":
				post = append(post, nb)
			default:
				other = append(other, nb)
			}
		}
		reordered := append(pre, append(post, other...)...)
		for _, nb := range nested {
			body.RemoveBlock(nb)
		}
		for _, nb := range reordered {
			body.AppendBlock(nb)
		}
	}

	return reorderBlock(block, order)
}

func init() { Register(outputStrategy{}) }
