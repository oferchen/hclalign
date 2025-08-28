// internal/ci/covercheck/doc.go

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
