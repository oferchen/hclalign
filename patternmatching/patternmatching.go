// patternmatching.go
// Provides functionality to compile glob patterns into regex and match files against these patterns.

package patternmatching

import (
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

// CompilePatterns compiles glob patterns into regular expressions for file matching.
func CompilePatterns(criteria []string) ([]*regexp.Regexp, error) {
	var compiledPatterns []*regexp.Regexp
	for _, globPattern := range criteria {
		regexPattern := translateGlobToRegex(globPattern)
		compiledPattern, err := regexp.Compile(regexPattern)
		if err != nil {
			return nil, err
		}
		compiledPatterns = append(compiledPatterns, compiledPattern)
	}
	return compiledPatterns, nil
}

// MatchesFileCriteria checks if the file name matches any of the compiled regex patterns.
func MatchesFileCriteria(filePath string, compiledPatterns []*regexp.Regexp) bool {
	baseName := filepath.Base(filePath)
	for _, pattern := range compiledPatterns {
		if pattern.MatchString(baseName) {
			return true
		}
	}
	return false
}

// translateGlobToRegex translates a glob pattern into a regex pattern.
func translateGlobToRegex(glob string) string {
	// Escape special characters, then replace glob patterns with regex equivalents.
	escaped := regexp.QuoteMeta(glob)
	regex := strings.ReplaceAll(escaped, "\\*", ".*")
	regex = strings.ReplaceAll(regex, "\\?", ".")
	return "^" + regex + "$"
}

// IsValidCriteria checks if each criterion in the criteria slice is valid.
// A valid criterion can be a specific filename (e.g., "main.tf"), a wildcard pattern (e.g., "*.tf"),
// or a directory pattern (with or without trailing slash).
func IsValidCriteria(criteria []string) bool {
	// This regex checks for:
	// - Wildcard patterns like "*.tf"
	// - Specific filenames like "main.tf"
	// - Directory patterns, which may end with a slash or have no extension
	validPattern := regexp.MustCompile(`^(\*|[a-zA-Z0-9_-]+)(\.[a-zA-Z0-9]+)?(/)?$`)

	for _, criterion := range criteria {
		if criterion == "" {
			log.Printf("Invalid criterion found: Criterion is empty")
			return false
		}
		if !validPattern.MatchString(criterion) {
			log.Printf("Invalid criterion found: %s", criterion)
			return false
		}
	}
	return true
}
