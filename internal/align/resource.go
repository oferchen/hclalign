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

	reqSet := map[string]struct{}{}
	optSet := map[string]struct{}{}
	compSet := map[string]struct{}{}
	var unk []string
	for _, name := range rest {
		if _, ok := opts.Schema.Required[name]; ok {
			reqSet[name] = struct{}{}
			continue
		}
		if _, ok := opts.Schema.Optional[name]; ok {
			optSet[name] = struct{}{}
			continue
		}
		if _, ok := opts.Schema.Computed[name]; ok {
			compSet[name] = struct{}{}
			continue
		}
		unk = append(unk, name)
	}

	var req, opt, comp []string
	for _, n := range opts.Schema.RequiredOrder {
		if _, ok := reqSet[n]; ok {
			req = append(req, n)
		}
	}
	for _, n := range opts.Schema.OptionalOrder {
		if _, ok := optSet[n]; ok {
			opt = append(opt, n)
		}
	}
	for _, n := range opts.Schema.ComputedOrder {
		if _, ok := compSet[n]; ok {
			comp = append(comp, n)
		}
	}

	order := append(metaAttrs, req...)
	order = append(order, opt...)
	order = append(order, comp...)
	order = append(order, unk...)

	return reorderBlock(block, order)
}
