import pyautogui
import time
import sys
import random

def simulate_mouse_movement(duration=5):
    """
    Simulasi gerakan mouse selama durasi yang ditentukan (dalam detik)
    """
    print(f"Simulating mouse movement for {duration} seconds...")
    
    # Dapatkan resolusi layar
    screen_width, screen_height = pyautogui.size()
    
    # Waktu mulai
    start_time = time.time()
    
    # Bergerak selama durasi yang ditentukan
    while time.time() - start_time < duration:
        # Posisi acak
        x = random.randint(100, screen_width - 100)
        y = random.randint(100, screen_height - 100)
        
        # Gerakan mouse ke posisi acak
        pyautogui.moveTo(x, y, duration=0.5)
        
        # Tunggu sedikit
        time.sleep(0.1)
    
    print("Mouse movement simulation completed!")

if __name__ == "__main__":
    # Default durasi 5 detik
    duration = 5
    
    # Jika ada argumen, gunakan sebagai durasi
    if len(sys.argv) > 1:
        try:
            duration = int(sys.argv[1])
        except ValueError:
            print(f"Invalid duration: {sys.argv[1]}. Using default: 5 seconds.")
    
    # Tunggu sedikit sebelum mulai simulasi
    time.sleep(1)
    
    # Jalankan simulasi
    simulate_mouse_movement(duration)