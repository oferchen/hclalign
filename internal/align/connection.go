// /internal/align/connection.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type connectionStrategy struct{}

func (connectionStrategy) Name() string { return "connection" }

func (connectionStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	allowed := map[string]struct{}{
		"type":                 {},
		"host":                 {},
		"user":                 {},
		"password":             {},
		"private_key":          {},
		"certificate":          {},
		"timeout":              {},
		"script_path":          {},
		"port":                 {},
		"proxy_command":        {},
		"agent":                {},
		"agent_socket":         {},
		"bastion_host":         {},
		"bastion_port":         {},
		"bastion_user":         {},
		"bastion_password":     {},
		"bastion_private_key":  {},
		"bastion_certificate":  {},
		"bastion_agent":        {},
		"bastion_agent_socket": {},
		"host_key":             {},
		"bastion_host_key":     {},
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

func init() { Register(connectionStrategy{}) }
