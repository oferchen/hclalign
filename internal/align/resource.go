// internal/align/resource.go
package align

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type resourceStrategy struct{}

func (resourceStrategy) Name() string { return "resource" }

func (resourceStrategy) Align(block *hclwrite.Block, opts *Options) error {
	return schemaAwareOrder(block, opts)
}

func init() { Register(resourceStrategy{}) }

func schemaAwareOrder(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	if opts == nil || opts.Schema == nil {
		sort.Strings(names)
		return reorderBlock(block, names)
	}

	var req, opt, comp, meta, unk []string
	for _, name := range names {
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
		if _, ok := opts.Schema.Meta[name]; ok {
			meta = append(meta, name)
			continue
		}
		unk = append(unk, name)
	}
	sort.Strings(req)
	sort.Strings(opt)
	sort.Strings(comp)
	sort.Strings(meta)
	sort.Strings(unk)

	if opts.Strict {
		for r := range opts.Schema.Required {
			if _, ok := attrs[r]; !ok {
				return fmt.Errorf("missing required attribute %q", r)
			}
		}
		if len(unk) > 0 {
			return fmt.Errorf("unknown attributes: %v", strings.Join(unk, ", "))
		}
	}

	final := append([]string{}, req...)
	final = append(final, opt...)
	final = append(final, comp...)
	final = append(final, meta...)
	final = append(final, unk...)
	return reorderBlock(block, final)
}
