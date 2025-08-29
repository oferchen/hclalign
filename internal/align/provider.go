// internal/align/provider.go
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
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))

	// Ensure alias is first if present
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
