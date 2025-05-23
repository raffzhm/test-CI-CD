package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

const (
	repoOwner = "pomokit"
	repoName  = "pomodoro"
	checkURL  = "https://api.github.com/repos/%s/%s/releases/latest"
)

type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

func isNewer(current, latest string) bool {
	// Hapus prefix 'v' jika ada
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	
	// Split versi berdasarkan titik
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")
	
	// Bandingkan setiap bagian
	for i := 0; i < len(currentParts) && i < len(latestParts); i++ {
		// Parse ke integer
		currentNum, _ := strconv.Atoi(currentParts[i])
		latestNum, _ := strconv.Atoi(latestParts[i])
		
		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
	}
	
	// Jika semua bagian sama, periksa jumlah bagian
	return len(latestParts) > len(currentParts)
}

func CheckForUpdates() {
	fmt.Println("🔍 Memeriksa pembaruan...")
	
	url := fmt.Sprintf(checkURL, repoOwner, repoName)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ Gagal membuat request: %v\n", err)
		return
	}
	
	req.Header.Set("User-Agent", "Pomokit-Version-Checker")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Gagal memeriksa pembaruan: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// Jika status code bukan 200 OK
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ GitHub API mengembalikan status code: %d\n", resp.StatusCode)
		return
	}
	
	// Parse response JSON
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("❌ Gagal parse response: %v\n", err)
		return
	}
	
	// Bersihkan versi tag (hapus 'v' di depan jika ada)
	latestVersionStr := strings.TrimPrefix(release.TagName, "v")
	currentVersionStr := Version
	
	// Bandingkan versi menggunakan fungsi sederhana
	if isNewer(currentVersionStr, latestVersionStr) {
		msg := fmt.Sprintf("Versi baru tersedia: %s (Anda menggunakan %s)\nMembuka halaman unduhan...", 
			release.TagName, Version)
		
		fmt.Println("✅ " + msg)
		
		// Tampilkan notifikasi
		beeep.Alert("Pomokit Update", msg, "information.png")
		
		// Langsung buka browser ke halaman unduhan tanpa menanyakan konfirmasi
		openbrowser(release.HTMLURL)
	} else {
		fmt.Println("✓ Anda sudah menggunakan versi terbaru")
	}
}