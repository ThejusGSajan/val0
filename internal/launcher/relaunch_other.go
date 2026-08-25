//go:build !windows

package launcher

// IsStandaloneConhost always returns false on non-Windows platforms.
func IsStandaloneConhost() bool {
	return false
}

// TryRelaunchInWT is a no-op on non-Windows platforms.
func TryRelaunchInWT(args []string) bool {
	return false
}
