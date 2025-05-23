//go:build windows

package main

import (
    "os"
    "syscall"
    "unsafe"
)

var quickEditDisabled = false

// Fungsi untuk mengaktifkan/menonaktifkan QuickEdit Mode
func setQuickEditMode(enable bool) {
    // Skip jika sudah dinonaktifkan dan permintaan untuk menonaktifkan lagi
    if !enable && quickEditDisabled {
        return
    }
    
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    procGetConsoleMode := kernel32.NewProc("GetConsoleMode")
    procSetConsoleMode := kernel32.NewProc("SetConsoleMode")
    
    var mode uint32
    handle := syscall.Handle(os.Stdin.Fd())
    
    procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
    
    if enable {
        // Enable QuickEdit mode (bit 0x0040)
        mode |= 0x0040
    } else {
        // Disable QuickEdit mode (clear bit 0x0040)
        mode &^= 0x0040
        quickEditDisabled = true
    }
    
    procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
}