// internal/align/lifecycle.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type lifecycleStrategy struct{}

func (lifecycleStrategy) Name() string { return "lifecycle" }

func (lifecycleStrategy) Align(block *hclwrite.Block, opts *Options) error {
	body := block.Body()
	attrs := body.Attributes()
	allowed := map[string]struct{}{
		"create_before_destroy": {},
		"prevent_destroy":       {},
		"ignore_changes":        {},
		"replace_triggered_by":  {},
	}
	names := make([]string, 0, len(attrs))
	for _, name := range []string{
		"create_before_destroy",
		"prevent_destroy",
		"ignore_changes",
		"replace_triggered_by",
	} {
		if _, ok := attrs[name]; ok {
			names = append(names, name)
		}
	}
	original := ihcl.AttributeOrder(body, attrs)
	for _, name := range original {
		if _, ok := allowed[name]; !ok {
			names = append(names, name)
		}
	}
	return reorderBlock(block, names)
}

func init() { Register(lifecycleStrategy{}) }
