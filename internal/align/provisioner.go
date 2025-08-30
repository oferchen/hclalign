// internal/align/provisioner.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type provisionerStrategy struct{}

func (provisionerStrategy) Name() string { return "provisioner" }

func (provisionerStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	allowed := map[string]struct{}{"when": {}, "on_failure": {}}
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

func init() { Register(provisionerStrategy{}) }
