package main

import (
	"fmt"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	panicIf(embd.InitGPIO())
	defer embd.CloseGPIO()

	fmt.Println("Creating new red LED.")
	red := NewLED(17)

	fmt.Println("Blinking pattern one for 10 seconds...")
	red.Blink(100*time.Millisecond, 1000*time.Millisecond)
	time.Sleep(10 * time.Second)

	fmt.Println("Blinking pattern two for 10 seconds...")
	red.Blink(500 * time.Millisecond)
	time.Sleep(10 * time.Second)

	fmt.Println("Done.")
	red.Stop()
	red.Stop()
}
