// /internal/align/provider.go
package align

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type providerStrategy struct{}

func (providerStrategy) Name() string { return "provider" }

func (providerStrategy) Align(block *hclwrite.Block, opts *Options) error {
	body := block.Body()

	nested := body.Blocks()
	if len(nested) > 0 {
		for _, nb := range nested {
			body.RemoveBlock(nb)
		}
		sort.Slice(nested, func(i, j int) bool {
			if nested[i].Type() == nested[j].Type() {
				li := strings.Join(nested[i].Labels(), "\x00")
				lj := strings.Join(nested[j].Labels(), "\x00")
				return li < lj
			}
			return nested[i].Type() < nested[j].Type()
		})
		for _, nb := range nested {
			body.AppendBlock(nb)
		}
	}

	attrs := body.Attributes()
	names := make([]string, 0, len(attrs))

	if _, ok := attrs["alias"]; ok {
		names = append(names, "alias")
	}

	others := make([]string, 0, len(attrs))
	for name := range attrs {
		if name == "alias" {
			continue
		}
		others = append(others, name)
	}
	sort.Strings(others)

	if opts != nil && opts.Strict && len(others) > 0 {
		return fmt.Errorf("provider: unknown attributes: %s", strings.Join(others, ", "))
	}

	names = append(names, others...)

	return reorderBlock(block, names)
}

func init() { Register(providerStrategy{}) }
