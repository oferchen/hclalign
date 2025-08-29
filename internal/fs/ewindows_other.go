// internal/fs/ewindows_other.go
//go:build !windows

package fs

func isErrWindows(err error) bool {
	return false
}
