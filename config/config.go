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
	Target      string
	Mode        Mode
	Stdin       bool
	Stdout      bool
	Include     []string
	Exclude     []string
	Order       []string
	StrictOrder bool
	Concurrency int
	Verbose     bool
}

// DefaultInclude, DefaultExclude and DefaultOrder define the default behaviour of the CLI.
var (
	DefaultInclude = []string{"**/*.tf"}
	DefaultExclude = []string{"**/.terraform/**", "**/vendor/**"}
	DefaultOrder   = []string{"description", "type", "default", "sensitive", "nullable", "validation"}
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

// ValidateOrder checks whether the provided order is valid. When strict is true
// the provided order must include all attributes in DefaultOrder exactly once.
// Otherwise it only checks for duplicate attributes.
func ValidateOrder(order []string, strict bool) error {
	providedSet := make(map[string]struct{})
	for _, item := range order {
		if _, exists := providedSet[item]; exists {
			return fmt.Errorf("duplicate attribute '%s' found in order", item)
		}
		providedSet[item] = struct{}{}
	}
	if strict {
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
