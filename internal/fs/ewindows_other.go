//go:build !windows

// internal/fs/ewindows_other.go — SPDX-License-Identifier: Apache-2.0
package fs

func isErrWindows(err error) bool {
	return false
}
