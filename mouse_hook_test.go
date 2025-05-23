package main

import (
    "fmt"
    "testing"
    "time"
)

func TestMouseHookDetection(t *testing.T) {
    fmt.Println("============ MOUSE HOOK TEST ============")
    fmt.Println("Initializing mouse hook...")
    
    err := initMouseHookDLL()
    if err != nil {
        t.Fatalf("Failed to initialize mouse hook: %v", err)
    }
    
    fmt.Println("Is hook installed:", IsHookInstalled())
    
    // Cek status setelah force activation
    ForceMouseActive()
    active := IsMouseHardwareActive(10 * time.Second)
    fmt.Printf("Mouse active after force activation: %v (expected: true)\n", active)
    
    if !active {
        t.Error("Mouse should be active after force activation")
    }
    
    fmt.Println("\nPlease move your physical mouse now...")
    fmt.Println("This will verify real mouse detection...")
    
    // Beri waktu user untuk menggerakkan mouse fisik
    time.Sleep(5 * time.Second)
    
    // Periksa apakah mouse aktif
    active = IsMouseHardwareActive(10 * time.Second)
    fmt.Printf("Mouse active after movement: %v\n", active)
    
    if !active {
        t.Error("Expected mouse to be active after movement")
    }
    
    fmt.Println("Now please do NOT move your mouse for 6 seconds...")
    
    // Tunggu 6 detik tanpa aktivitas
    time.Sleep(6 * time.Second)
    
    // Periksa dengan timeout 5 detik
    stillActive := IsMouseHardwareActive(5 * time.Second)
    fmt.Printf("Mouse still active after 6s inactivity with 5s timeout: %v (expected: false)\n", stillActive)
    
    // Periksa dengan timeout 10 detik
    stillActiveWithLongerTimeout := IsMouseHardwareActive(10 * time.Second)
    fmt.Printf("Mouse still active with 10s timeout: %v (expected: true)\n", stillActiveWithLongerTimeout)
    
    fmt.Println("============================================")
}