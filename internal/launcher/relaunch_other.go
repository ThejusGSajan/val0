//go:build !windows

package launcher

import "os/exec"

// IsStandaloneConhost always returns false on non-Windows platforms.
func IsStandaloneConhost() bool {
	return false
}

// TryRelaunchInWT is a no-op on non-Windows platforms.
func TryRelaunchInWT(args []string) bool {
	return false
}

// FindWTPath always returns exec.ErrNotFound on non-Windows platforms.
func FindWTPath() (string, error) {
	return "", exec.ErrNotFound
}
