//go:build darwin

package main

// setQuickEditMode - macOS implementation (no-op)
// QuickEdit mode is a Windows-specific console feature
func setQuickEditMode(enable bool) {
    // No-op on macOS - QuickEdit mode doesn't exist
    // This function is left empty intentionally
    // Console behavior on macOS is handled by the terminal emulator
}