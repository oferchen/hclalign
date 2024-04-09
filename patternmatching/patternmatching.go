// patternmatching.go
// Provides functionality to compile glob patterns into regex and match files against these patterns.

package patternmatching

import (
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

