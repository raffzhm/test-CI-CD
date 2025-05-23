package main

import (
    "fmt"
    "os/exec"
    "testing"
    "time"
)

func TestMouseHookVsPythonScript(t *testing.T) {
    fmt.Println("============ ANTI-CHEAT TEST ============")
    fmt.Println("Initializing mouse hook...")
    
    err := initMouseHookDLL()
    if err != nil {
        t.Fatalf("Failed to initialize mouse hook: %v", err)
    }
    
    // Reset status mouse
    fmt.Println("Waiting for 5 seconds of inactivity to reset state...")
    time.Sleep(5 * time.Second)
    
    // Verifikasi status sebelum simulasi
    beforeScript := IsMouseHardwareActive(3 * time.Second)
    fmt.Printf("Mouse active before script (after 5s inactivity): %v (expected: false)\n", beforeScript)
    
    // Jalankan skrip Python untuk simulasi mouse
    fmt.Println("\nRunning Python script to simulate mouse movement...")
    cmd := exec.Command("python", "mouse_simulator.py", "3")
    
    // Tampilkan output skrip
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("Failed to run Python script: %v\nOutput: %s", err, output)
    }
    
    fmt.Printf("Python script output: %s\n", output)
    
    // Cek status setelah simulasi Python
    afterScript := IsMouseHardwareActive(3 * time.Second)
    fmt.Printf("Mouse active after Python script: %v\n", afterScript)
    
    // Jika hook kita bekerja baik, seharusnya TIDAK mendeteksi simulasi Python
    // sebagai gerakan mouse hardware fisik
    if afterScript {
        fmt.Println("WARNING: Hook detected Python-simulated mouse movement as hardware movement!")
        fmt.Println("This suggests the anti-cheat might not be fully effective.")
    } else {
        fmt.Println("SUCCESS: Hook correctly ignored Python-simulated mouse movement!")
        fmt.Println("This confirms the anti-cheat is working as intended.")
    }
    
    // Sekarang minta user menggerakkan mouse fisik
    fmt.Println("\nNow please move your PHYSICAL mouse...")
    time.Sleep(5 * time.Second)
    
    // Cek status setelah gerakan fisik
    afterPhysical := IsMouseHardwareActive(3 * time.Second)
    fmt.Printf("Mouse active after physical movement: %v (expected: true)\n", afterPhysical)
    
    // Keberhasilan penuh jika bisa membedakan gerakan fisik dan simulasi
    if !afterScript && afterPhysical {
        fmt.Println("\nPERFECT! The system correctly distinguishes between:")
        fmt.Println("✓ Physical mouse movement (detected)")
        fmt.Println("✓ Simulated movement (ignored)")
        fmt.Println("\nYour anti-cheat system is working perfectly!")
    } else if !afterScript && !afterPhysical {
        fmt.Println("\nPARTIAL SUCCESS: The system ignores simulated movements,")
        fmt.Println("but also fails to detect physical movements correctly.")
    } else if afterScript && afterPhysical {
        fmt.Println("\nPARTIAL FAILURE: The system detects physical movements,")
        fmt.Println("but also incorrectly detects simulated movements as real.")
    } else {
        fmt.Println("\nFAILURE: The system's behavior is unexpected.")
    }
    
    fmt.Println("============================================")
}