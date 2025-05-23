//go:build windows

package main

import (
    "fmt"
    "syscall"
    "time"
)

var (
    mouseHookDLL       *syscall.DLL
    procInitHook       *syscall.Proc
    procIsActive       *syscall.Proc
    procForceActive    *syscall.Proc
    procIsHookInstalled *syscall.Proc
    hookLoaded         bool
    hookInitError      error
)

// Inisialisasi DLL dan prosedurnya
func initMouseHookDLL() error {
    if hookLoaded {
        return hookInitError
    }
    
    // Load DLL
    var err error
    mouseHookDLL, err = syscall.LoadDLL("MouseHook.dll")
    if err != nil {
        hookInitError = fmt.Errorf("failed to load MouseHook.dll: %v", err)
        return hookInitError
    }
    
    // Dapatkan pointer ke prosedur
    procInitHook, err = mouseHookDLL.FindProc("InitMouseHook")
    if err != nil {
        hookInitError = fmt.Errorf("failed to find InitMouseHook procedure: %v", err)
        return hookInitError
    }
    
    procIsActive, err = mouseHookDLL.FindProc("IsMouseActive")
    if err != nil {
        hookInitError = fmt.Errorf("failed to find IsMouseActive procedure: %v", err)
        return hookInitError
    }
    
    procForceActive, err = mouseHookDLL.FindProc("ForceMouseActive")
    if err != nil {
        hookInitError = fmt.Errorf("failed to find ForceMouseActive procedure: %v", err)
        return hookInitError
    }
    
    procIsHookInstalled, err = mouseHookDLL.FindProc("IsHookInstalled")
    if err != nil {
        hookInitError = fmt.Errorf("failed to find IsHookInstalled procedure: %v", err)
        return hookInitError
    }
    
    // Inisialisasi hook
    ret, _, _ := procInitHook.Call()
    if ret == 0 {
        hookInitError = fmt.Errorf("failed to initialize mouse hook")
        return hookInitError
    }
    
    // Verifikasi hook terpasang
    ret, _, _ = procIsHookInstalled.Call()
    if ret == 0 {
        hookInitError = fmt.Errorf("mouse hook reported as not installed after initialization")
        return hookInitError
    }
    
    // Aktifkan untuk testing
    procForceActive.Call()
    
    hookLoaded = true
    hookInitError = nil
    fmt.Println("")
    return nil
}

// IsMouseHardwareActive mengecek apakah mouse hardware aktif dalam interval waktu tertentu
func IsMouseHardwareActive(timeout time.Duration) bool {
    if err := initMouseHookDLL(); err != nil {
        fmt.Printf("Warning: %v - Falling back to simple detection\n", err)
        return false
    }
    
    // Konversi timeout ke milliseconds
    timeoutMs := uint32(timeout.Milliseconds())
    
    // Panggil fungsi DLL
    ret, _, _ := procIsActive.Call(uintptr(timeoutMs))
    
    return ret != 0
}

// ForceMouseActive memaksa status mouse menjadi aktif
func ForceMouseActive() {
    if err := initMouseHookDLL(); err != nil {
        fmt.Printf("Warning: %v\n", err)
        return
    }
    
    procForceActive.Call()
}

// IsHookInstalled memeriksa apakah hook berhasil dipasang
func IsHookInstalled() bool {
    if err := initMouseHookDLL(); err != nil {
        return false
    }
    
    ret, _, _ := procIsHookInstalled.Call()
    return ret != 0
}