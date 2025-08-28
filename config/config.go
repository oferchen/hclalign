// config.go
// Manages configuration settings and validation for HCL file alignment.

package config

import (
	"context"
	"fmt"
	"github.com/oferchen/hclalign/fileprocessing"
	"github.com/oferchen/hclalign/patternmatching"
	"strings"
)

// Config stores configuration for processing HCL files.
type Config struct {
	Target   string
	Criteria []string
	Order    []string
}

// DefaultCriteria and DefaultOrder define the default behavior of the CLI.
var (
	DefaultCriteria = []string{"*.tf"}
	DefaultOrder    = []string{"description", "type", "default", "sensitive", "nullable", "validation"}
)

const (
	MissingTarget = "missing target file or directory. Please provide a valid target as an argument"
)

// IsValidOrder checks if the provided order is valid, returning an error with specific feedback if not.
func IsValidOrder(order []string) (bool, error) {
	defaultOrder := DefaultOrder
	providedSet := make(map[string]struct{})

	for _, item := range order {
		if _, exists := providedSet[item]; exists {
			// Duplicate item found, invalid order
			return false, fmt.Errorf("duplicate attribute '%s' found in order", item)
		}
		providedSet[item] = struct{}{}
	}

	if len(providedSet) != len(defaultOrder) {
		// Provided order doesn't match the default order's length
		return false, fmt.Errorf("provided order length %d doesn't match expected %d", len(providedSet), len(defaultOrder))
	}

	for _, item := range defaultOrder {
		if _, exists := providedSet[item]; !exists {
			// An item from the defaultOrder is not in the provided order
			return false, fmt.Errorf("missing expected attribute '%s' in provided order", item)
		}
	}

	// All checks passed, valid order
	return true, nil
}

// ProcessTargetDynamically processes files in the target directory based on criteria and order.
func ProcessTargetDynamically(ctx context.Context, target string, criteria []string, order []string) error {
	if err := patternmatching.IsValidCriteria(criteria); err != nil {
		return fmt.Errorf("invalid criteria: %w", err)
	}
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("no target specified")
	}

	ctx = context.WithValue(ctx, fileprocessing.TargetContextKey, target)
	return fileprocessing.ProcessFiles(ctx, target, criteria, order)
}
