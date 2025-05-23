package main

import (
	"image"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/kbinani/screenshot"
)

func TakeScreenshot() {
	n := screenshot.NumActiveDisplays()

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}
		ScreenShotStack = append(ScreenShotStack, img)
	}
}

func ImageToFile(img *image.RGBA) (filename string) {
	id := uuid.New()
	filename = id.String() + ".png"
	file, _ := os.Create(filename)
	defer file.Close()
	png.Encode(file, img)
	return
}

func GetRandomScreensot(ScreenShootStack []*image.RGBA) (img *image.RGBA) {
	randomIndex := rand.Intn(len(ScreenShootStack))
	img = ScreenShootStack[randomIndex]
	return
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func StringtoFile(content string, filepath string) {
	data := []byte(content)
	err := os.WriteFile(filepath, data, 0)
	if err != nil {
		log.Fatal(err)
	}
}

func FiletoString(filepath string) (content string) {
	if FileExists(filepath) {
		bytecontent, err := os.ReadFile(filepath)
		if err != nil {
			log.Println(err)
		} else {
			content = string(bytecontent)
		}
	}
	return

}
