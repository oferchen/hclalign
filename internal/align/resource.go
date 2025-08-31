// internal/align/resource.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"

	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type resourceStrategy struct{}

func (resourceStrategy) Name() string { return "resource" }

func (resourceStrategy) Align(block *hclwrite.Block, opts *Options) error {
	return schemaAwareOrder(block, opts)
}

func init() { Register(resourceStrategy{}) }

func schemaAwareOrder(block *hclwrite.Block, opts *Options) error {
	body := block.Body()
	attrs := body.Attributes()
	originalOrder := ihcl.AttributeOrder(body, attrs)

	canonical := CanonicalBlockAttrOrder[block.Type()]
	metaAttrs := make([]string, 0, len(canonical))
	metaSet := map[string]struct{}{}
	for _, n := range canonical {
		if _, ok := attrs[n]; ok {
			metaAttrs = append(metaAttrs, n)
			metaSet[n] = struct{}{}
		}
	}

	var rest []string
	for _, n := range originalOrder {
		if _, ok := metaSet[n]; !ok {
			rest = append(rest, n)
		}
	}

	if opts == nil || opts.Schema == nil {
		order := append(metaAttrs, rest...)
		return reorderBlock(block, order)
	}

	var req, opt, comp, unk []string
	for _, name := range rest {
		if _, ok := opts.Schema.Required[name]; ok {
			req = append(req, name)
			continue
		}
		if _, ok := opts.Schema.Optional[name]; ok {
			opt = append(opt, name)
			continue
		}
		if _, ok := opts.Schema.Computed[name]; ok {
			comp = append(comp, name)
			continue
		}
		unk = append(unk, name)
	}

	order := append(metaAttrs, req...)
	order = append(order, opt...)
	order = append(order, comp...)
	order = append(order, unk...)

	return reorderBlock(block, order)
}
