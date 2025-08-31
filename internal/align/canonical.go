// internal/align/canonical.go
package align

var CanonicalBlockAttrOrder = map[string][]string{
	"variable": {"description", "type", "default", "sensitive", "nullable"},
	"output":   {"description", "value", "sensitive", "ephemeral", "depends_on"},
	"module":   {"source", "version", "providers", "count", "for_each", "depends_on"},
	"provider": {"alias"},
	"resource": {"provider", "count", "for_each", "depends_on"},
	"data":     {"provider", "count", "for_each", "depends_on"},
	"terraform": {
		"required_version",
		"required_providers",
		"backend",
		"cloud",
	},
}
