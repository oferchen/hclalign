// config.go
// Manages configuration settings and validation for HCL file alignment.

package config

import (
	"fmt"

	"github.com/oferchen/hclalign/patternmatching"
)

// Mode represents the operation mode of the application.
type Mode int

const (
	// ModeWrite writes the formatted content back to the files.
	ModeWrite Mode = iota
	// ModeCheck only checks if formatting changes are required.
	ModeCheck
	// ModeDiff prints the diff of required changes.
	ModeDiff
)

// Config stores configuration for processing HCL files.
type Config struct {
	Target         string
	Mode           Mode
	Stdin          bool
	Stdout         bool
	Include        []string
	Exclude        []string
	Order          []string
	StrictOrder    bool
	Concurrency    int
	Verbose        bool
	FollowSymlinks bool
}

// DefaultInclude, DefaultExclude and DefaultOrder define the default behaviour of the CLI.
var (
	DefaultInclude = []string{"**/*.tf"}
	DefaultExclude = []string{"**/.terraform/**", "**/vendor/**", "**/.git/**", "**/node_modules/**"}
	// DefaultOrder defines the canonical ordering of variable block attributes.
	// Nested blocks such as "validation" are always appended after attributes
	// and therefore are not part of this list.
	DefaultOrder = []string{"description", "type", "default", "sensitive", "nullable"}
)

const (
	// MissingTarget is returned when no target is specified and --stdin is not used.
	MissingTarget = "missing target file or directory. Please provide a valid target as an argument"
)

// Validate ensures that the configuration is valid.
func (c *Config) Validate() error {
	if c.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}
	if err := patternmatching.ValidatePatterns(c.Include); err != nil {
		return fmt.Errorf("invalid include: %w", err)
	}
	if err := patternmatching.ValidatePatterns(c.Exclude); err != nil {
		return fmt.Errorf("invalid exclude: %w", err)
	}
	if err := ValidateOrder(c.Order, c.StrictOrder); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}
	return nil
}

// ValidateOrder checks whether the provided order is valid. Duplicate
// attributes always cause an error. When strict is true, all attributes must be
// from the canonical DefaultOrder list and each must appear exactly once. The
// canonical list only contains attribute names; nested blocks like "validation"
// are not considered part of the order.
func ValidateOrder(order []string, strict bool) error {
	providedSet := make(map[string]struct{})
	canonicalSet := make(map[string]struct{}, len(DefaultOrder))
	for _, item := range DefaultOrder {
		canonicalSet[item] = struct{}{}
	}

	for _, item := range order {
		if _, exists := providedSet[item]; exists {
			return fmt.Errorf("duplicate attribute '%s' found in order", item)
		}
		providedSet[item] = struct{}{}
	}

	if strict {
		for item := range providedSet {
			if _, ok := canonicalSet[item]; !ok {
				return fmt.Errorf("unknown attribute '%s' in order", item)
			}
		}
		if len(providedSet) != len(DefaultOrder) {
			return fmt.Errorf("provided order length %d doesn't match expected %d", len(providedSet), len(DefaultOrder))
		}
		for _, item := range DefaultOrder {
			if _, exists := providedSet[item]; !exists {
				return fmt.Errorf("missing expected attribute '%s' in provided order", item)
			}
		}
	}
	return nil
}
