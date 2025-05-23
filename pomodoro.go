package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
)

func main() {

	CheckForUpdates()
	initMouseHookDLL()
	setQuickEditMode(true)

	fmt.Println("Pomokit " + Version)
	fmt.Println("\nSetelah menautkan akun whatsapp, mohon untuk tidak melepas tautan whatsmeow selama sesi pomodoro")
	fmt.Println("Lalu pastikan terhubung ke internet selama sesi pomodoro")

	rand.Seed(time.Now().UnixNano())

	if !FileExists("information.png") {
		go DownloadFile("information.png", InfoImageURL)
	}
	if !FileExists("warning.png") {
		go DownloadFile("warning.png", WarningImageURL)
	}

	wag := InputWAGroup()
	InputURLGithub()
	milestone := InputMilestone()

	WhatsApp()
	SendNotifTo(wag, milestone)

	GetSetTime("task")
	GetSetTime("break")

	GetSetTime("task")
	GetSetTime("break")

	GetSetTime("task")
	GetSetTime("break")

	GetSetTime("task")
	GetSetTime("break")

	GetSetTime("longbreak")

	img := GetRandomScreensot(ScreenShotStack)
	filename := ImageToFile(img)

	if time.Since(tokenCreationTime) > (2 * time.Hour) {
		RefreshToken() // Memperbarui currentHashURL
	}

	SendReportTo(filename, wag, milestone, currentHashURL)
	msg := "Selamat!!!!! 1 sesi pomodoro selesai dengan jumlah skrinsutan:" + strconv.Itoa(len(ScreenShotStack))

	beeep.Alert("Pomokit Info", msg, "information.png")
	fmt.Println("1 sesi pomodoro selesai dengan jumlah skrinsutan:", strconv.Itoa(len(ScreenShotStack)))
	fmt.Println("Tekan Ctrl+C untuk keluar aplikasi")
	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	WAclient.Disconnect()

}
