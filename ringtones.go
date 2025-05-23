package main

import (
	"time"

	"github.com/gen2brain/beeep"
)

func nokiaTune() {
	beeep.Beep(1319, 133)
	time.Sleep(143 * time.Millisecond)
	beeep.Beep(1175, 133)
	time.Sleep(143 * time.Millisecond)
	beeep.Beep(740, 267)
	time.Sleep(277 * time.Millisecond)
	beeep.Beep(831, 267)
	time.Sleep(277 * time.Millisecond)
	beeep.Beep(1109, 133)
	time.Sleep(143 * time.Millisecond)
	beeep.Beep(988, 133)
	time.Sleep(143 * time.Millisecond)
	beeep.Beep(587, 267)
	time.Sleep(277 * time.Millisecond)
	beeep.Beep(659, 267)
	time.Sleep(277 * time.Millisecond)
	beeep.Beep(988, 133)
	time.Sleep(143 * time.Millisecond)
	beeep.Beep(880, 133)
	time.Sleep(143 * time.Millisecond)
	beeep.Beep(554, 267)
	time.Sleep(277 * time.Millisecond)
	beeep.Beep(659, 267)
	time.Sleep(277 * time.Millisecond)
	beeep.Beep(880, 533)
	time.Sleep(543 * time.Millisecond)

}
