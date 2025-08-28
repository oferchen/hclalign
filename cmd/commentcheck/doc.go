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
=======

// Package main provides the commentcheck command. It verifies that each Go
// source file starts with a comment containing its repository-relative path.
//
// Example
//
//	$ cat hello.go
//	// hello.go
//	package hello
//
//	$ commentcheck
//	(no output)
//
// If the comment is missing:
//
//	$ cat bad.go
//	package hello
//
//	$ commentcheck
//	bad.go: first line must be "// bad.go"

package main
