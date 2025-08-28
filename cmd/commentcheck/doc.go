// cmd/commentcheck/doc.go
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
// If the comment is missing or incorrect:
//
//	$ cat bad.go
//	package hello
//
//	$ commentcheck
//	bad.go: first line must be "// bad.go"
package main
