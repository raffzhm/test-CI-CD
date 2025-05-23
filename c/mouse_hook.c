#include <windows.h>
#include <stdio.h>

// Struktur untuk menyimpan status mouse hardware
typedef struct {
    BOOL isActive;
    DWORD lastActivityTime;
    BOOL isHookInstalled;
} MouseState;

static MouseState mouseState = {FALSE, 0, FALSE};
static HHOOK hMouseHook = NULL;
static HINSTANCE hInstance = NULL;

// Hook procedure untuk Low-Level Mouse Hook
LRESULT CALLBACK LowLevelMouseProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode == HC_ACTION) {
        // Deteksi jika mouse bergerak/klik fisik
        switch (wParam) {
            case WM_MOUSEMOVE:
            case WM_LBUTTONDOWN:
            case WM_LBUTTONUP:
            case WM_RBUTTONDOWN:
            case WM_RBUTTONUP:
            case WM_MBUTTONDOWN:
            case WM_MBUTTONUP:
            case WM_MOUSEWHEEL:
                // Update status aktivitas mouse
                mouseState.isActive = TRUE;
                mouseState.lastActivityTime = GetTickCount();
                break;
        }
    }
    
    // Lanjutkan ke hook berikutnya
    return CallNextHookEx(hMouseHook, nCode, wParam, lParam);
}

// Thread untuk hook mouse
DWORD WINAPI MouseHookThread(LPVOID lpParam) {
    MSG msg;
    
    // Install hook mouse low-level
    hMouseHook = SetWindowsHookEx(WH_MOUSE_LL, LowLevelMouseProc, hInstance, 0);
    
    if (!hMouseHook) {
        return 1;
    }
    
    mouseState.isHookInstalled = TRUE;
    
    // Loop pesan untuk hook
    while (GetMessage(&msg, NULL, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
    
    // Cleanup hook jika loop pesan selesai
    UnhookWindowsHookEx(hMouseHook);
    mouseState.isHookInstalled = FALSE;
    
    return 0;
}

// Inisialisasi hook mouse
__declspec(dllexport) BOOL InitMouseHook() {
    hInstance = GetModuleHandle(NULL);
    
    if (!hInstance) {
        return FALSE;
    }
    
    // Setup hook di thread terpisah
    HANDLE hThread = CreateThread(NULL, 0, MouseHookThread, NULL, 0, NULL);
    
    if (!hThread) {
        return FALSE;
    }
    
    // Tunggu sampai hook terpasang atau timeout
    DWORD startTime = GetTickCount();
    while (!mouseState.isHookInstalled) {
        Sleep(10);
        if (GetTickCount() - startTime > 2000) { // 2 detik timeout
            return FALSE;
        }
    }
    
    CloseHandle(hThread);
    return TRUE;
}

// Cek apakah mouse aktif dalam periode waktu tertentu
__declspec(dllexport) BOOL IsMouseActive(DWORD timeoutMs) {
    if (!mouseState.isHookInstalled) {
        return FALSE;
    }
    
    if (!mouseState.isActive) {
        return FALSE;
    }
    
    // Periksa apakah aktivitas terakhir masih dalam batas waktu
    DWORD currentTime = GetTickCount();
    DWORD elapsedTime;
    
    // Tangani overflow (GetTickCount akan overflow setelah ~49 hari)
    if (currentTime < mouseState.lastActivityTime) {
        // Overflow terjadi
        elapsedTime = (0xFFFFFFFF - mouseState.lastActivityTime) + currentTime;
    } else {
        elapsedTime = currentTime - mouseState.lastActivityTime;
    }
    
    return elapsedTime <= timeoutMs;
}

// Mengaktifkan deteksi mouse untuk testing
__declspec(dllexport) void ForceMouseActive() {
    mouseState.isActive = TRUE;
    mouseState.lastActivityTime = GetTickCount();
}

// Mendapatkan status hook
__declspec(dllexport) BOOL IsHookInstalled() {
    return mouseState.isHookInstalled;
}

// Entry point untuk DLL
BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
    switch (fdwReason) {
        case DLL_PROCESS_ATTACH:
            hInstance = hinstDLL;
            break;
            
        case DLL_PROCESS_DETACH:
            // Cleanup resources
            if (hMouseHook) {
                UnhookWindowsHookEx(hMouseHook);
                hMouseHook = NULL;
            }
            break;
    }
    
    return TRUE;
}