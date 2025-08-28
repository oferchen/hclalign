// cmd/commentcheck/doc.go

// Command commentcheck verifies that each Go source file in the module
// starts with a comment containing its repository-relative path.
//
// Example:
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
