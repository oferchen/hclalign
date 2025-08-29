// internal/align/dynamic.go
package align

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type dynamicStrategy struct{}

func (dynamicStrategy) Name() string { return "dynamic" }

func (dynamicStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	allowed := map[string]struct{}{"for_each": {}, "iterator": {}}
	var unknown []string
	for name := range attrs {
		names = append(names, name)
		if _, ok := allowed[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	if opts != nil && opts.Strict && len(unknown) > 0 {
		sort.Strings(unknown)
		return fmt.Errorf("dynamic: unknown attributes: %s", strings.Join(unknown, ", "))
	}
	sort.Strings(names)
	return reorderBlock(block, names)
}

func init() { Register(dynamicStrategy{}) }
