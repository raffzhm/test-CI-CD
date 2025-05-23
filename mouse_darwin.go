//go:build darwin

package main

import (
    "sync"
    "time"
    "github.com/go-vgo/robotgo"
)

var (
    // lastMouseX       int
    // lastMouseY       int
    // lastActivityTime time.Time
    mouseMutex       sync.RWMutex
    mouseInitialized bool
)

// initMouseHookDLL - macOS implementation (no DLL needed)
func initMouseHookDLL() error {
    mouseMutex.Lock()
    defer mouseMutex.Unlock()
    
    if !mouseInitialized {
        // Initialize mouse tracking for macOS
        lastMouseX, lastMouseY = robotgo.GetMousePos()
        lastActivityTime = time.Now()
        mouseInitialized = true
        
        // Start background mouse monitoring
        go monitorMouseActivity()
    }
    
    return nil
}

// Background goroutine to monitor mouse activity on macOS
func monitorMouseActivity() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for range ticker.C {
        mouseMutex.Lock()
        currentX, currentY := robotgo.GetMousePos()
        
        if currentX != lastMouseX || currentY != lastMouseY {
            lastMouseX = currentX
            lastMouseY = currentY
            lastActivityTime = time.Now()
        }
        mouseMutex.Unlock()
    }
}

// IsMouseHardwareActive mengecek apakah mouse hardware aktif dalam interval waktu tertentu (macOS)
func IsMouseHardwareActive(timeout time.Duration) bool {
    if err := initMouseHookDLL(); err != nil {
        return false
    }
    
    mouseMutex.RLock()
    defer mouseMutex.RUnlock()
    
    // Check if mouse has been active within the timeout period
    return time.Since(lastActivityTime) <= timeout
}

// ForceMouseActive memaksa status mouse menjadi aktif (macOS)
func ForceMouseActive() {
    if err := initMouseHookDLL(); err != nil {
        return
    }
    
    mouseMutex.Lock()
    defer mouseMutex.Unlock()
    
    lastActivityTime = time.Now()
}

// IsHookInstalled memeriksa apakah hook berhasil dipasang (macOS)
func IsHookInstalled() bool {
    mouseMutex.RLock()
    defer mouseMutex.RUnlock()
    
    return mouseInitialized
}