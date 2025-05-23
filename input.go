package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/whatsauth/watoken"
)

func CheckURLStatus(url string) (status bool, msg string) {
	fmt.Println("\n🔍 Memulai verifikasi URL...")

	if !ValidUrl(url) {
		fmt.Println("❌ Format URL tidak valid")
		return
	}

	fmt.Println("⏳ Mengecek Tipe Url...")

	if strings.Contains(url, "gtmetrix.com") {
		fmt.Println("⏳ Memulai Ambil data GTmetrix...")
		data, err := ScrapeGTmetrixData(url)
		if err == nil && len(data) > 0 {
			fmt.Println("✅ Berhasil memverifikasi")
			status = true
			msg = "GTmetrix URL valid (verified through scraping)"
			return
		} else {
			fmt.Printf("❌ Gagal : %v\n", err)
			msg = "GTmetrix URL could not be scraped: " + err.Error()
			return
		}
	}

	fmt.Println("🌐 Mengecek ketersediaan URL...")
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Gagal koneksi: %v\n", err)
		msg = err.Error()
		return
	}
	defer response.Body.Close()

	fmt.Printf("📡 Status response: %s\n", response.Status)
	msg = response.Status
	status = (msg == "200 OK")
	return
}

func ValidUrl(urllink string) bool {
	_, er := url.Parse(urllink)
	return er == nil
}

func InputWAGroup() (wag string) {
	beeep.Alert("Pomokit Info", "Please Input Your WhatsApp Group ID with keyword : Myika minta id grup : ", "information.png")
	fmt.Println("\nPlease Input Your WhatsApp Group ID with keyword : Myika minta id grup : ")
	fmt.Scanln(&wag)
	return
}

var OriginalURL string          // Variabel global untuk menyimpan URL asli
var tokenCreationTime time.Time // Waktu pembuatan token
var currentHashURL string       // Token URL terkini
var PrivateKey = "a68a14331baa7e777b194c97daa86b987882f7d9b4b7291eef5099f13134cbcd622fde17d02151da5e69fffd01e356cafb09fa28f316246dfe88d65a2fee28f1"

func InputURLGithub() (hashurl string) {
	var urltask string
	fmt.Println("input URL Yang Akan Dikerjakan(copas dari browser) : ")
	fmt.Scanln(&urltask)
	urlvalid, msgerrurl := CheckURLStatus(urltask)
	for !urlvalid {
		beeep.Alert("Invalid Link", "URL Tidak Valid : "+msgerrurl, "information.png")
		fmt.Println("URL Invalid, Masukkan kembali URL yang benar : ")
		fmt.Scanln(&urltask)
		urlvalid, msgerrurl = CheckURLStatus(urltask)
	}
	OriginalURL = urltask // Simpan URL asli di variabel global

	var alias = urltask
	hashurl, err := watoken.EncodeforHours(urltask, alias, PrivateKey, 3)
	if err != nil {
		fmt.Println(err)
	}

	// Simpan informasi token baru
	currentHashURL = hashurl
	// fmt.Println("Token URL baru:", hashurl)
	tokenCreationTime = time.Now()

	return
}

func InputMilestone() (milestone string) {
	beeep.Alert("Pomokit Info", "Silahkan input rencana yang akan anda kerjakan pada 1 cycle pomodoro sekarang", "information.png")
	fmt.Println("\nRencana yang akan anda kerjakan pada 1 cycle pomodoro sekarang : ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		milestone = scanner.Text()
		if len(milestone) > 17 {
			break
		} else {
			beeep.Alert("Pomokit Info", "Rencana belum diisi atau terlalu pendek", "information.png")
			fmt.Println("Rencana belum diisi atau terlalu pendek, Rencana Anda : ")
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return

}
