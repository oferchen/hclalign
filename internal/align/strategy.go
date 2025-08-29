package align

import "github.com/hashicorp/hcl/v2/hclwrite"

// Options control how strategies behave.
type Options struct {
	// Order defines preferred attribute ordering for certain strategies.
	Order []string
	// Strict enforces strict ordering rules.
	Strict bool
}

// Strategy defines alignment behaviour for a particular block type.
type Strategy interface {
	// Name returns block type handled by this strategy.
	Name() string
	// Align mutates the given block according to strategy rules.
	Align(block *hclwrite.Block, opts *Options) error
}

var registry = map[string]Strategy{}

// Register adds s to the registry keyed by its name.
func Register(s Strategy) {
	registry[s.Name()] = s
}

// Apply walks blocks in the given file and applies registered strategies.
func Apply(file *hclwrite.File, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	return applyBody(file.Body(), opts)
}

func applyBody(body *hclwrite.Body, opts *Options) error {
	for _, b := range body.Blocks() {
		if strat, ok := registry[b.Type()]; ok {
			if err := strat.Align(b, opts); err != nil {
				return err
			}
		}
		if err := applyBody(b.Body(), opts); err != nil {
			return err
		}
	}
	return nil
}

// ReorderAttributes is a backwards compatible helper used by existing code.
// It simply delegates to Apply with Options.
func ReorderAttributes(file *hclwrite.File, order []string, strict bool) error {
	return Apply(file, &Options{Order: order, Strict: strict})
}
