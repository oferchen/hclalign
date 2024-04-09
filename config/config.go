// config.go
// Manages configuration settings and validation for HCL file alignment.

package config

import (
	"fmt"
	"strings"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/fileprocessing"
	"github.com/oferchen/hclalign/patternmatching"
)

// Config stores configuration for processing HCL files.
type Config struct {
	Target   string
	Criteria []string
	Order    []string
}

// IsValidOrder checks if the provided order is valid, returning an error with specific feedback if not.
func IsValidOrder(order []string) (bool, error) {
	defaultOrder := cli.DefaultOrder
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
