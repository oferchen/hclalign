// filename: internal/align/lifecycle.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type lifecycleStrategy struct{}

func (lifecycleStrategy) Name() string { return "lifecycle" }

func (lifecycleStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	allowed := map[string]struct{}{
		"create_before_destroy": {},
		"prevent_destroy":       {},
		"ignore_changes":        {},
		"replace_triggered_by":  {},
	}
	var known, unknown []string
	for name := range attrs {
		if _, ok := allowed[name]; ok {
			known = append(known, name)
		} else {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(known)
	sort.Strings(unknown)
	names := append(known, unknown...)
	return reorderBlock(block, names)
}

func init() { Register(lifecycleStrategy{}) }
