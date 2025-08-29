package align

import "github.com/hashicorp/hcl/v2/hclwrite"

// noopStrategy is used for block types where alignment is not yet implemented.
type noopStrategy struct{ name string }

func (n noopStrategy) Name() string { return n.name }

func (noopStrategy) Align(_ *hclwrite.Block, _ *Options) error { return nil }

func init() {
	// Register basic strategies for block types for future extension.
	names := []string{
		"output",
		"locals",
		"module",
		"provider",
		"terraform",
		"resource",
		"data",
		"dynamic",
		"lifecycle",
		"provisioner",
	}
	for _, n := range names {
		Register(noopStrategy{name: n})
	}
}
