//go:build !windows
// internal/fs/ewindows_other.go
package fs

func isErrWindows(err error) bool {
	return false
}
