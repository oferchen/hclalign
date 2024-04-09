// config.go
// Manages configuration settings and validation for HCL file alignment.

package config

import (
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

// IsValidOrder checks if the provided order is valid.
func IsValidOrder(order []string) bool {
	defaultOrder := []string{"description", "type", "default", "sensitive", "nullable", "validation"}
	providedSet := make(map[string]struct{})
	for _, item := range order {
		if _, exists := providedSet[item]; exists {
			return false // Duplicate item found, invalid order
		}
		providedSet[item] = struct{}{}
	}

	if len(providedSet) != len(defaultOrder) {
		return false // Ensures the provided order has the exact number of items as the default order
	}

	for _, item := range defaultOrder {
		if _, exists := providedSet[item]; !exists {
			return false // An item from defaultOrder is not in the provided order, invalid order
		}
	}

	return true // All checks passed, valid order
}

// ProcessTargetDynamically processes files in the target directory based on criteria and order.
func ProcessTargetDynamically(target string, criteria []string, order []string) error {
	// Implementation assumes existence of fileprocessing and patternmatching packages
	if !patternmatching.IsValidCriteria(criteria) {
		return fmt.Errorf("invalid criteria: %v", criteria)
	}
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("no target specified")
	}

	// Further processing logic to be implemented in fileprocessing.ProcessFiles
	return fileprocessing.ProcessFiles(target, criteria, order)
}

