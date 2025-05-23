package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

// IsGTmetrixURL mengecek apakah URL berasal dari laporan GTmetrix
func IsGTmetrixURL(url string) bool {
	return strings.Contains(url, "gtmetrix.com/reports/")
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",
}

// ScrapeGTmetrixWithChromedp menggunakan browser otomatis untuk mengambil data GTmetrix
func ScrapeGTmetrixWithChromedp(url string) (map[string]string, error) {
	fmt.Println("⏳ Mencoba scraping dengan browser otomatis...")
	performanceData := make(map[string]string)

	// Buat context dengan timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Opsi untuk browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"),
	)

	// Buat allocator context
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Buat browser context
	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Jalankan tasks
	var htmlContent string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(url),
		// Tambahkan delay agar terlihat seperti manusia
		chromedp.Sleep(3*time.Second),
		// Scroll ke bawah secara bertahap (seperti manusia)
		chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight/4, behavior: 'smooth'})`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight/2, behavior: 'smooth'})`, nil),
		chromedp.Sleep(1*time.Second),
		// Ambil HTML
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		return nil, fmt.Errorf("error otomatisasi browser: %v", err)
	}

	// Deteksi apakah laporan kedaluwarsa
	if hasExpiredIndicators(htmlContent) {
		fmt.Println("⚠️ Terdeteksi laporan GTmetrix kedaluwarsa!")
		return nil, fmt.Errorf("laporan GTmetrix sudah kedaluwarsa dan tidak dapat diakses. Silakan gunakan URL laporan GTmetrix yang masih berlaku")
	}

	// === TAMBAHAN BARU: Ambil URL asli website yang diuji ===
	var originalURL string
	_ = chromedp.Run(browserCtx,
		// Coba ambil dari tag h2 > a.no-external yang biasanya berisi URL asli
		chromedp.AttributeValue(`h2 > a.no-external`, "href", &originalURL, nil),
	)

	// Jika berhasil mendapatkan URL
	if originalURL != "" {
		// fmt.Println("  ✓ Berhasil mendapatkan URL asli website:", originalURL)
		performanceData["OriginalURL"] = originalURL
	}

	// Jika URL belum didapat, coba dari meta description
	if originalURL == "" {
		var metaDesc string
		_ = chromedp.Run(browserCtx,
			chromedp.AttributeValue(`meta[name="description"]`, "content", &metaDesc, nil),
		)
		
		// Ekstrak URL dari meta description
		if strings.Contains(metaDesc, "Latest Performance Report for: ") {
			originalURL = strings.TrimPrefix(metaDesc, "Latest Performance Report for: ")
			// fmt.Println("  ✓ Berhasil mendapatkan URL asli website dari meta:", originalURL)
			performanceData["OriginalURL"] = originalURL
		}
	}

	// === PERBAIKAN UTAMA: Ambil meta tags property og:description ===
	var metaOGDescription string
	_ = chromedp.Run(browserCtx,
		chromedp.AttributeValue(`meta[property="og:description"]`, "content", &metaOGDescription, nil),
	)

	if metaOGDescription != "" {
		fmt.Println("  ✓ Berhasil mendapatkan meta og:description")

		// Parse og:description - biasanya dalam format "GTmetrix Grade: B (Performance: 86% / Structure: 79%)"
		gradeRegex := regexp.MustCompile(`GTmetrix Grade: ([A-F])`)
		if matches := gradeRegex.FindStringSubmatch(metaOGDescription); len(matches) > 1 {
			performanceData["Grade"] = matches[1]
		}

		perfRegex := regexp.MustCompile(`Performance: (\d+)%`)
		if matches := perfRegex.FindStringSubmatch(metaOGDescription); len(matches) > 1 {
			performanceData["Performance"] = matches[1] + "%"
		}

		structRegex := regexp.MustCompile(`Structure: (\d+)%`)
		if matches := structRegex.FindStringSubmatch(metaOGDescription); len(matches) > 1 {
			performanceData["Structure"] = matches[1] + "%"
		}
	}

	// === PERBAIKAN TAMBAHAN: Ambil nilai langsung dari elemen score di halaman ===
	var grade, performance, structure, lcp, tbt, cls string

	// Ambil grade dengan lebih spesifik
	_ = chromedp.Run(browserCtx,
		// Selector untuk Grade yang akurat (target kelas icon specific)
		chromedp.TextContent(`.report-score-grade-gtmetrix .icon-grade-[A-F]`, &grade, chromedp.ByQuery),
	)

	// Jika selector spesifik gagal, coba selector alternatif
	if grade == "" {
		_ = chromedp.Run(browserCtx,
			chromedp.Evaluate(`document.querySelector('.icon-grade-A, .icon-grade-B, .icon-grade-C, .icon-grade-D, .icon-grade-E, .icon-grade-F')?.classList[0]?.replace('icon-grade-', '')`, &grade),
		)
	}

	// Ambil Performance dan Structure dengan selector class yang tepat
	_ = chromedp.Run(browserCtx,
		chromedp.TextContent(`.color-grade-B .report-score-percent`, &performance, chromedp.ByQuery),
		chromedp.TextContent(`.color-grade-C .report-score-percent`, &structure, chromedp.ByQuery),
	)

	// Ambil Web Vitals - perhatikan bahwa selector .report-web-vital mengacu pada elemen-elemen dalam section Web Vitals
	_ = chromedp.Run(browserCtx,
		chromedp.TextContent(".report-web-vital:nth-child(1) .report-web-vital-value", &lcp, chromedp.ByQuery),
		chromedp.TextContent(".report-web-vital:nth-child(2) .report-web-vital-value", &tbt, chromedp.ByQuery),
		chromedp.TextContent(".report-web-vital:nth-child(3) .report-web-vital-value", &cls, chromedp.ByQuery),
	)

	// Masukkan data yang berhasil diambil ke map
	if grade != "" && strings.TrimSpace(grade) != "" {
		// Ambil hanya karakter pertama untuk Grade - cleanup jika diperlukan
		grade = strings.TrimSpace(grade)
		if len(grade) > 0 {
			performanceData["Grade"] = string(grade[0])
		}
	}

	if performance != "" && strings.TrimSpace(performance) != "" {
		performanceData["Performance"] = strings.TrimSpace(performance)
	}

	if structure != "" && strings.TrimSpace(structure) != "" {
		performanceData["Structure"] = strings.TrimSpace(structure)
	}

	if lcp != "" && strings.TrimSpace(lcp) != "" {
		performanceData["LCP"] = strings.TrimSpace(lcp)
	}

	if tbt != "" && strings.TrimSpace(tbt) != "" {
		performanceData["TBT"] = strings.TrimSpace(tbt)
	}

	if cls != "" && strings.TrimSpace(cls) != "" {
		performanceData["CLS"] = strings.TrimSpace(cls)
	}

	// === PERBAIKAN ALTERNATIF: Ekstrak dari JavaScript data ===
	// Cek apakah Performance atau Structure belum ada
	_, perfExists := performanceData["Performance"]
	_, structExists := performanceData["Structure"]

	if !perfExists || !structExists {
		var jsVitalsData string
		_ = chromedp.Run(browserCtx,
			// Coba ambil nilai dari variabel JavaScript di halaman
			chromedp.Evaluate(`
				(function() {
					if (window.GTmetrix && window.GTmetrix.vars) {
						if (window.GTmetrix.vars.vitals) {
							return JSON.stringify(window.GTmetrix.vars.vitals);
						}
					}
					return "";
				})()
			`, &jsVitalsData),
		)

		if jsVitalsData != "" {
			fmt.Println("  ✓ Berhasil mendapatkan data dari JavaScript variables")

			// Cari Performance score jika belum ada
			if !perfExists {
				perfScoreRegex := regexp.MustCompile(`"performance_score":(\d+(\.\d+)?)`)
				if matches := perfScoreRegex.FindStringSubmatch(jsVitalsData); len(matches) > 1 {
					// Konversi ke persentase integer
					perfScore, err := strconv.ParseFloat(matches[1], 64)
					if err == nil {
						performanceData["Performance"] = fmt.Sprintf("%d%%", int(perfScore*100))
					}
				}
			}

			// Cari Structure score jika belum ada
			if !structExists {
				structScoreRegex := regexp.MustCompile(`"structure_score":(\d+(\.\d+)?)`)
				if matches := structScoreRegex.FindStringSubmatch(jsVitalsData); len(matches) > 1 {
					// Konversi ke persentase integer
					structScore, err := strconv.ParseFloat(matches[1], 64)
					if err == nil {
						performanceData["Structure"] = fmt.Sprintf("%d%%", int(structScore*100))
					}
				}
			}
		}
	}

	// === FALLBACK KE REGEX yang lebih permisif jika selector gagal ===
	// Grade (jika belum ada)
	if _, exists := performanceData["Grade"]; !exists {
		gradeRegex := regexp.MustCompile(`icon-grade-([A-F])`)
		if matches := gradeRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			performanceData["Grade"] = matches[1]
		}
	}

	// Performance (jika belum ada)
	if _, exists := performanceData["Performance"]; !exists {
		perfPatterns := []string{
			`<span class="report-score-grade color-grade-[A-F]"><span class="report-score-percent">(\d+)%</span></span>`,
			`Performance.*?<span class="report-score-percent">(\d+)%</span>`,
			`Performance.*?(\d+)%`,
		}

		for _, pattern := range perfPatterns {
			perfRegex := regexp.MustCompile(pattern)
			if matches := perfRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
				performanceData["Performance"] = matches[1] + "%"
				break
			}
		}
	}

	// Structure (jika belum ada)
	if _, exists := performanceData["Structure"]; !exists {
		structPatterns := []string{
			`<span class="report-score-grade color-grade-[A-F]"><span class="report-score-percent">(\d+)%</span></span>`,
			`Structure.*?<span class="report-score-percent">(\d+)%</span>`,
			`Structure.*?(\d+)%`,
		}

		for _, pattern := range structPatterns {
			structRegex := regexp.MustCompile(pattern)
			if matches := structRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
				if _, perfExists := performanceData["Performance"]; !perfExists || matches[1] != strings.TrimSuffix(performanceData["Performance"], "%") {
					// Pastikan ini bukan nilai Performance yang sama
					performanceData["Structure"] = matches[1] + "%"
					break
				}
			}
		}
	}

	// Web Vitals (LCP, TBT, CLS)
	if _, exists := performanceData["LCP"]; !exists {
		lcpRegex := regexp.MustCompile(`<span class="report-web-vital-value[^"]*">([0-9.]+s)</span>`)
		if matches := lcpRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			performanceData["LCP"] = matches[1]
		}
	}

	if _, exists := performanceData["TBT"]; !exists {
		tbtRegex := regexp.MustCompile(`<span class="report-web-vital-value[^"]*">(0ms|[0-9.]+ms)</span>`)
		if matches := tbtRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			performanceData["TBT"] = matches[1]
		}
	}

	if _, exists := performanceData["CLS"]; !exists {
		clsRegex := regexp.MustCompile(`<span class="report-web-vital-value[^"]*">([0-9.]+)</span>`)
		if matches := clsRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			performanceData["CLS"] = matches[1]
		}
	}

	// Cetak semua data yang berhasil diambil
	for key, value := range performanceData {
		fmt.Printf("  ✓ %s: %s\n", key, value)
	}

	return performanceData, nil
}

// ScrapeGTmetrixData dimodifikasi untuk menggunakan colly terlebih dahulu, kemudian chromedp sebagai fallback
func ScrapeGTmetrixData(url string) (map[string]string, error) {
	fmt.Println("⏳ Memulai proses rekap data GTmetrix dengan Metode pertama...")
	fmt.Println("⚠️ Proses ini mungkin memakan waktu 15-30 detik")

	performanceData := make(map[string]string)

	// Menggunakan userAgents yang sudah ada dari kode asli
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",
	}

	// Persiapkan collector Colly
	c := colly.NewCollector(
		colly.UserAgent(userAgents[rand.Intn(len(userAgents))]),
		colly.AllowURLRevisit(),
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// Konfigurasi transport
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
		Proxy:             http.ProxyFromEnvironment,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	})

	// Header lengkap
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("Referer", "https://gtmetrix.com/")
		r.Headers.Set("DNT", "1")
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "same-origin")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	})

	// Cookie handling
	c.SetCookies("gtmetrix.com", []*http.Cookie{
		{Name: "cookie_consent", Value: "true"},
		{Name: "gtm_session", Value: "1"},
	})

	// Limit rule
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*gtmetrix.*",
		Parallelism: 1,
		RandomDelay: 15 * time.Second,
		Delay:       10 * time.Second,
	})

	// Flag untuk mendeteksi laporan kedaluwarsa
	reportExpired := false

	// Deteksi laporan kedaluwarsa
	c.OnHTML("body", func(e *colly.HTMLElement) {
		htmlContent := e.Text
		if hasExpiredIndicators(htmlContent) {
			reportExpired = true
		}
	})

	// Error handling dengan retry
	retryCount := 0
	maxRetries := 2
	collyFailed := false

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request failed: %s (Status: %d)", r.Request.URL, r.StatusCode)

		if r.StatusCode == 403 && retryCount < maxRetries {
			retryCount++
			waitTime := time.Duration(rand.Intn(120)+60) * time.Second
			fmt.Printf("⏱️ Menunggu %v sebelum retry ke-%d\n", waitTime, retryCount)

			time.Sleep(waitTime)
			err = r.Request.Retry()
			if err != nil {
				fmt.Println("🔴 Retry gagal:", err)
				collyFailed = true
			}
		} else {
			collyFailed = true
		}
	})

	// ============== TAMBAHKAN PENGAMBILAN URL ASLI ==============
	// Ambil URL asli dari title atau meta description
	c.OnHTML("title", func(e *colly.HTMLElement) {
		title := e.Text
		if strings.Contains(title, "Latest Performance Report for:") {
			// Format: "Latest Performance Report for: https://pomokit.github.io/pomodoro/ | GTmetrix"
			parts := strings.Split(title, "Latest Performance Report for:")
			if len(parts) > 1 {
				urlPart := strings.Split(parts[1], "|")[0]
				urlPart = strings.TrimSpace(urlPart)
				if urlPart != "" {
					// fmt.Println("✅ Berhasil mendapatkan URL asli dari title:", urlPart)
					performanceData["OriginalURL"] = urlPart
				}
			}
		}
	})

	c.OnHTML("meta[name='description']", func(e *colly.HTMLElement) {
		content := e.Attr("content")
		// fmt.Println("Found meta description:", content)
		
		// Check if it contains a URL
		if strings.Contains(content, "Latest Performance Report for:") {
			urlPart := strings.TrimPrefix(content, "Latest Performance Report for: ")
			urlPart = strings.TrimSpace(urlPart)
			if urlPart != "" {
				// fmt.Println("✅ Berhasil mendapatkan URL asli dari meta description:", urlPart)
				performanceData["OriginalURL"] = urlPart
			}
		}
	})

	// Try to find the URL directly from the h2 link which usually contains the tested URL
	c.OnHTML("h2 a.no-external", func(e *colly.HTMLElement) {
		urlLink := e.Attr("href")
		if urlLink != "" {
			// fmt.Println("✅ Berhasil mendapatkan URL asli dari link:", urlLink)
			performanceData["OriginalURL"] = urlLink
		}
	})
	// ============== END PENGAMBILAN URL ASLI ==============

	// ============== PARSING LOGIC ASLI ==============
	c.OnHTML("meta[property='og:description']", func(e *colly.HTMLElement) {
		content := e.Attr("content")
		fmt.Println("Found meta description:", content)

		gradeRegex := regexp.MustCompile(`GTmetrix Grade: ([A-F])`)
		if matches := gradeRegex.FindStringSubmatch(content); len(matches) > 1 {
			performanceData["Grade"] = matches[1]
		}

		perfRegex := regexp.MustCompile(`Performance: (\d+)%`)
		if matches := perfRegex.FindStringSubmatch(content); len(matches) > 1 {
			performanceData["Performance"] = matches[1] + "%"
		}

		structRegex := regexp.MustCompile(`Structure: (\d+)%`)
		if matches := structRegex.FindStringSubmatch(content); len(matches) > 1 {
			performanceData["Structure"] = matches[1] + "%"
		}
	})

	c.OnHTML(".report-score-grade-gtmetrix .icon-grade-A, .report-score-grade-gtmetrix i[class*='icon-grade-']", func(e *colly.HTMLElement) {
		className := e.Attr("class")
		gradeRegex := regexp.MustCompile(`icon-grade-([A-F])`)
		if matches := gradeRegex.FindStringSubmatch(className); len(matches) > 1 {
			performanceData["Grade"] = matches[1]
		}
	})

	c.OnHTML(".report-score-grade.color-grade-B .report-score-percent, .report-score h4:contains('Performance') + span", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		performanceData["Performance"] = text
	})

	c.OnHTML(".report-score-grade.color-grade-A .report-score-percent, .report-score h4:contains('Structure') + span", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		performanceData["Structure"] = text
	})

	c.OnHTML(".report-web-vital-value.color-rating-med-low, .report-web-vital h4:contains('LCP') + span", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		performanceData["LCP"] = text
	})

	c.OnHTML(".report-web-vital-value.color-rating-low, .report-web-vital h4:contains('TBT') + span", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if strings.Contains(text, "ms") || strings.Contains(text, "0ms") {
			performanceData["TBT"] = text
		}
	})

	c.OnHTML(".report-web-vital-value.color-rating-med, .report-web-vital h4:contains('CLS') + span", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if text != "" && !strings.Contains(text, "s") && !strings.Contains(text, "ms") {
			performanceData["CLS"] = text
		}
	})

	c.OnHTML(".report-web-vital", func(e *colly.HTMLElement) {
		heading := e.DOM.Find("h4").Text()
		value := e.DOM.Find(".report-web-vital-value").Text()
		value = strings.TrimSpace(value)

		if value == "" {
			return
		}

		if strings.Contains(heading, "Largest Contentful Paint") || strings.Contains(heading, "LCP") {
			performanceData["LCP"] = value
		} else if strings.Contains(heading, "Total Blocking Time") || strings.Contains(heading, "TBT") {
			performanceData["TBT"] = value
		} else if strings.Contains(heading, "Cumulative Layout Shift") || strings.Contains(heading, "CLS") {
			performanceData["CLS"] = value
		}
	})
	// ============== END PARSING LOGIC ==============

	// Eksekusi Colly
	err := c.Visit(url)
	if err != nil {
		collyFailed = true
		fmt.Printf("🔴 Metode pertama gagal mengakses: %v\n", err)
	}

	c.Wait()

	// Cek apakah laporan kedaluwarsa terdeteksi
	if reportExpired {
		return nil, fmt.Errorf("laporan GTmetrix sudah kedaluwarsa dan tidak dapat diakses. Silakan gunakan URL laporan GTmetrix yang masih berlaku")
	}

	// Jika Colly gagal atau performanceData kurang dari 3 item, coba dengan Chromedp
	if collyFailed || len(performanceData) < 3 {
		fmt.Println("🔄 Metode pertama gagal mendapatkan data yang cukup, mencoba dengan browser otomatis...")

		chromedpData, chromedpErr := ScrapeGTmetrixWithChromedp(url)
		if chromedpErr == nil && len(chromedpData) >= 3 {
			fmt.Println("✅ Berhasil mendapatkan data dengan browser otomatis")
			return chromedpData, nil
		}

		fmt.Printf("🔴 Browser otomatis juga gagal: %v\n", chromedpErr)

		// Jika keduanya gagal, coba scraper alternatif atau kembalikan data yang ada
		if len(performanceData) > 0 {
			return performanceData, nil
		}

		// Jika semua gagal dan tidak ada data sama sekali
		return nil, fmt.Errorf("gagal mendapatkan data GTmetrix dengan semua metode")
	}

	return performanceData, nil
}

func ScrapeLargerScope(url string) (map[string]string, error) {
	performanceData := make(map[string]string)

	c := colly.NewCollector(
		colly.UserAgent(userAgents[rand.Intn(len(userAgents))]),
	)

	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
		Proxy:             http.ProxyFromEnvironment,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Referer", "https://gtmetrix.com/")
		r.Headers.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		RandomDelay: 20 * time.Second,
	})

	retryCount := 0
	maxRetries := 1
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request failed: %s (Status: %d)", r.Request.URL, r.StatusCode)

		if r.StatusCode == 403 && retryCount < maxRetries {
			retryCount++
			waitTime := time.Duration(rand.Intn(120)+60) * time.Second
			log.Printf("Retry %d/%d in %v", retryCount, maxRetries, waitTime)

			time.Sleep(waitTime)
			err = r.Request.Retry()
			if err != nil {
				log.Println("Retry failed:", err)
			}
		}
	})

	// ============== PARSING LOGIC ASLI ==============
	c.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Println("Scanning full HTML page for performance metrics")

		e.ForEach("*", func(_ int, el *colly.HTMLElement) {
			text := strings.TrimSpace(el.Text)

			if len(text) == 1 && text >= "A" && text <= "F" {
				parentHTML, _ := el.DOM.Parent().Html()
				if strings.Contains(strings.ToLower(parentHTML), "grade") ||
					strings.Contains(strings.ToLower(parentHTML), "gtmetrix") {
					performanceData["Grade"] = text
				}
			}

			if strings.HasSuffix(text, "%") && len(text) <= 4 {
				parentText := el.DOM.Parent().Text()

				if strings.Contains(strings.ToLower(parentText), "performance") {
					performanceData["Performance"] = text
				} else if strings.Contains(strings.ToLower(parentText), "structure") {
					performanceData["Structure"] = text
				}
			}

			if strings.HasSuffix(text, "s") && len(text) <= 5 && !strings.HasSuffix(text, "ms") {
				parentText := el.DOM.Parent().Text()
				if strings.Contains(strings.ToLower(parentText), "lcp") ||
					strings.Contains(strings.ToLower(parentText), "largest contentful paint") {
					performanceData["LCP"] = text
				}
			}

			if strings.HasSuffix(text, "ms") && len(text) <= 5 {
				parentText := el.DOM.Parent().Text()
				if strings.Contains(strings.ToLower(parentText), "tbt") ||
					strings.Contains(strings.ToLower(parentText), "total blocking time") {
					performanceData["TBT"] = text
				}
			}

			if strings.Contains(text, ".") && len(text) <= 5 && !strings.Contains(text, "s") {
				parentText := el.DOM.Parent().Text()
				if strings.Contains(strings.ToLower(parentText), "cls") ||
					strings.Contains(strings.ToLower(parentText), "cumulative layout shift") {
					performanceData["CLS"] = text
				}
			}
		})
	})
	// ============== END PARSING LOGIC ==============

	err := c.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed after %d retries: %v", retryCount, err)
	}

	c.Wait()

	return performanceData, nil
}

// FallbackMessage memberikan pesan yang sesuai untuk nilai yang tidak ditemukan
func FallbackMessage(metric string) string {
	return "[data " + metric + " tidak tersedia]"
}

// FormatGTmetrixReport membuat pesan WhatsApp terformat dengan data GTmetrix
func FormatGTmetrixReport(data map[string]string, url string) string {
	if len(data) == 0 {
		return "\n\n*Data GTmetrix*\nTidak dapat mengambil data performa dari " + url
	}

	report := "\n\n*Rekap Data GTmetrix*"

	// Tambahkan URL yang dianalisis - KODE YANG DIMODIFIKASI
	if originalURL, exists := data["OriginalURL"]; exists {
		// Gunakan URL asli dari hasil scraping jika tersedia
		report += "\nWebsite: " + originalURL
	} else {
		// Fallback ke metode lama jika tidak menemukan OriginalURL
		urlParts := strings.Split(url, "/reports/")
		if len(urlParts) > 1 {
			domainParts := strings.Split(urlParts[1], "/")
			if len(domainParts) > 0 {
				report += "\nWebsite: " + domainParts[0]
			}
		}
	}

	// Tambahkan metrik dalam urutan tertentu
	if grade, ok := data["Grade"]; ok {
		report += "\nGrade: " + grade
	} else {
		report += "\nGrade: " + FallbackMessage("grade")
	}

	if perf, ok := data["Performance"]; ok {
		report += "\nPerformance: " + perf
	} else {
		report += "\nPerformance: " + FallbackMessage("performance")
	}

	if structure, ok := data["Structure"]; ok {
		report += "\nStructure: " + structure
	} else {
		report += "\nStructure: " + FallbackMessage("structure")
	}

	if lcp, ok := data["LCP"]; ok {
		report += "\nLCP (Largest Contentful Paint): " + lcp
	} else {
		report += "\nLCP (Largest Contentful Paint): " + FallbackMessage("LCP")
	}

	if tbt, ok := data["TBT"]; ok {
		report += "\nTBT (Total Blocking Time): " + tbt
	} else {
		report += "\nTBT (Total Blocking Time): " + FallbackMessage("TBT")
	}

	if cls, ok := data["CLS"]; ok {
		report += "\nCLS (Cumulative Layout Shift): " + cls
	} else {
		report += "\nCLS (Cumulative Layout Shift): " + FallbackMessage("CLS")
	}

	if len(data) < 4 {
		report += "\n\n*Catatan: Beberapa data tidak dapat diambil. Silakan buka URL GTmetrix untuk hasil yang lengkap.*"
	}

	return report
}

func hasExpiredIndicators(htmlContent string) bool {
	expiredIndicators := []string{
		"This report is expired",
		"report-truncated-placeholder",
		"filter-blur",
		"pending purge",
		"expired and pending purge",
	}
	
	for _, indicator := range expiredIndicators {
		if strings.Contains(htmlContent, indicator) {
			return true
		}
	}
	
	return false
}