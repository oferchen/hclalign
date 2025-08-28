// internal/ci/covercheck/doc.go
// Command covercheck reads a Go coverage profile and verifies that the
// total percentage meets a preset threshold (95% by default).
//
// Example usage:
//
//	$ go test -coverprofile=coverage.out ./...
//	$ covercheck
//	Total coverage: 96.0%
//
// If coverage falls below the threshold, covercheck exits with a non-zero status.
package main
