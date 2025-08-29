//go:build !windows

package fs

func isErrWindows(err error) bool {
	return false
}
