// internal/align/lifecycle.go
package align

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type lifecycleStrategy struct{}

func (lifecycleStrategy) Name() string { return "lifecycle" }

func (lifecycleStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	allowed := map[string]struct{}{
		"create_before_destroy": {},
		"prevent_destroy":       {},
		"ignore_changes":        {},
		"replace_triggered_by":  {},
	}
	var unknown []string
	for name := range attrs {
		names = append(names, name)
		if _, ok := allowed[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	if opts != nil && opts.Strict && len(unknown) > 0 {
		sort.Strings(unknown)
		return fmt.Errorf("lifecycle: unknown attributes: %s", strings.Join(unknown, ", "))
	}
	sort.Strings(names)
	return reorderBlock(block, names)
}

func init() { Register(lifecycleStrategy{}) }
