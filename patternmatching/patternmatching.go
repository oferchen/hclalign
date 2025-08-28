// patternmatching.go
// Provides functionality to validate and match file patterns using filepath.Match.

package patternmatching

import (
	"fmt"
	"path/filepath"
)

// IsValidCriteria checks if each criterion in the criteria slice is valid.
// It returns an error describing the first invalid criterion encountered.
func IsValidCriteria(criteria []string) error {
	for _, criterion := range criteria {
		if criterion == "" {
			return fmt.Errorf("invalid criterion: criterion is empty")
		}
		if _, err := filepath.Match(criterion, ""); err != nil {
			return fmt.Errorf("invalid criterion '%s': %w", criterion, err)
		}
	}
	return nil
}

// MatchesFileCriteria checks if the file name matches any of the provided glob patterns.
func MatchesFileCriteria(filePath string, criteria []string) bool {
	baseName := filepath.Base(filePath)
	for _, pattern := range criteria {
		matched, err := filepath.Match(pattern, baseName)
		if err == nil && matched {
			return true
		}
	}
	return false
}
