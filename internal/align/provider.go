// /internal/align/provider.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type providerStrategy struct{}

func (providerStrategy) Name() string { return "provider" }

func (providerStrategy) Align(block *hclwrite.Block, opts *Options) error {
	body := block.Body()

	attrs := body.Attributes()
	names := make([]string, 0, len(attrs))

	if _, ok := attrs["alias"]; ok {
		names = append(names, "alias")
	}

	order := ihcl.AttributeOrder(body, attrs)
	others := make([]string, 0, len(order))
	for _, name := range order {
		if name == "alias" {
			continue
		}
		others = append(others, name)
	}

	names = append(names, others...)

	return reorderBlock(block, names)
}

func init() { Register(providerStrategy{}) }
