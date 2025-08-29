// /internal/align/strategy.go
package align

import "github.com/hashicorp/hcl/v2/hclwrite"

type Options struct {
	Order      []string
	BlockOrder map[string]string

	Strict bool

	Schemas map[string]*Schema

	Schema *Schema
}

type Schema struct {
	Required map[string]struct{}
	Optional map[string]struct{}
	Computed map[string]struct{}
	Meta     map[string]struct{}
}

type Strategy interface {
	Name() string

	Align(block *hclwrite.Block, opts *Options) error
}

var registry = map[string]Strategy{}

func Register(s Strategy) {
	registry[s.Name()] = s
}

func Apply(file *hclwrite.File, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	return applyBody(file.Body(), opts)
}

func applyBody(body *hclwrite.Body, opts *Options) error {
	for _, b := range body.Blocks() {
		if strategy, ok := registry[b.Type()]; ok {
			sub := *opts
			sub.Schema = nil
			if len(b.Labels()) > 0 && opts.Schemas != nil {
				if s, ok := opts.Schemas[b.Labels()[0]]; ok {
					sub.Schema = s
				}
			}
			if err := strategy.Align(b, &sub); err != nil {
				return err
			}
		}
		if err := applyBody(b.Body(), opts); err != nil {
			return err
		}
	}
	return nil
}

func ReorderAttributes(file *hclwrite.File, order []string, strict bool) error {
	return Apply(file, &Options{Order: order, Strict: strict})
}
