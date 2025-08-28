// cmd/commentcheck/doc.go
// Command commentcheck verifies that each Go source file in the module starts with a
// comment matching its relative path.
//
// Example usage:
//
//	$ commentcheck
//	internal/foo/bar.go: first line must be "// internal/foo/bar.go"
//
// The tool exits with status 1 when mismatches are found.
package main
