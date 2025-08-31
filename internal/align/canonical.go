// internal/align/canonical.go
package align

import "github.com/oferchen/hclalign/config"

var CanonicalBlockAttrOrder = map[string][]string{
	"variable": append([]string(nil), config.CanonicalOrder...),
	"output":   {"description", "value", "sensitive", "ephemeral", "depends_on"},
	"module":   {"source", "version", "providers", "count", "for_each", "depends_on"},
	"provider": {"alias"},
	"resource": {"provider", "count", "for_each", "depends_on"},
	"data":     {"provider", "count", "for_each", "depends_on"},
}
