//go:build windows

package launcher

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
)

// IsStandaloneConhost checks if the current binary was launched as the sole process
// attached to the console host (typical when double-clicking in Explorer under legacy conhost.exe).
func IsStandaloneConhost() bool {
	// If already in Windows Terminal, WT_SESSION is set
	if os.Getenv("WT_SESSION") != "" {
		return false
	}

	// Prevent loop if explicitly launched via relauncher
	if os.Getenv("VAL0_RELAUNCHED") == "1" {
		return false
	}

	var pids [2]uint32
	count, _, _ := procGetConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids)),
	)
	// If count <= 1, this console was created exclusively for this process (Explorer launch)
	// If launched from an existing cmd.exe / powershell.exe, count >= 2
	return count <= 1
}

// FindWTPath locates wt.exe via PATH or the default WindowsApps directory.
func FindWTPath() (string, error) {
	if path, err := exec.LookPath("wt.exe"); err == nil {
		return path, nil
	}

	// Fallback to local AppData WindowsApps path
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		wtAppPath := filepath.Join(localAppData, "Microsoft", "WindowsApps", "wt.exe")
		if _, err := os.Stat(wtAppPath); err == nil {
			return wtAppPath, nil
		}
	}

	return "", exec.ErrNotFound
}

// TryRelaunchInWT attempts to relaunch the current executable inside Windows Terminal (wt.exe).
// Returns false if wt.exe cannot be found or if spawning fails.
func TryRelaunchInWT(args []string) bool {
	wtPath, err := FindWTPath()
	if err != nil {
		return false
	}

	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	// Construct command to launch inside Windows Terminal
	// wt.exe --title "val0" <exePath> <args...>
	wtArgs := append([]string{"--title", "val0", exePath}, args...)
	cmd := exec.Command(wtPath, wtArgs...)
	cmd.Env = append(os.Environ(), "VAL0_RELAUNCHED=1")

	if err := cmd.Start(); err == nil {
		// Exit the legacy conhost process so the user transitions smoothly into Windows Terminal
		os.Exit(0)
		return true
	}

	return false
}
