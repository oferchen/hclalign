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
=======

// Package main provides the covercheck command used in CI pipelines. It reads a
// Go coverage profile and fails if the total coverage drops below 95 percent.
//
// Example
//
//	$ go test -coverprofile=coverage.out ./...
//	$ covercheck
//	Total coverage: 97.3%
//
// When coverage is too low, the command exits with an error:
//
//	$ covercheck
//	Total coverage: 90.0%
//	Coverage 90.0% is below 95.0%

package main
