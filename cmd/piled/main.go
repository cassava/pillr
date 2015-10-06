// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/cassava/pillr/led"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	pinnr := flag.Int("pin", -1, "gpio pin number of LED")
	flag.Parse()

	if *pinnr < 0 {
		fmt.Println("Please specify pin which LED is on! Be careful!")
		os.Exit(1)
	}

	panicIf(embd.InitGPIO())
	defer embd.CloseGPIO()

	l := led.New(*pinnr)
	if flag.NArg() <= 0 {
		fmt.Println("Blinking Heartbeat1000 pattern till you quit...")
		l.Blink(led.Heartbeat1000...)
	} else {
		pattern := make([]time.Duration, flag.NArg())
		for i, s := range flag.Args() {
			d, err := time.ParseDuration(s)
			if err != nil {
				fmt.Println("Error parsing duration:", err)
				os.Exit(1)
			}
			pattern[i] = d
		}

		fmt.Printf("Blinking your pattern %v till you quit...\n", pattern)
		l.Blink(pattern...)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("\nBye-bye.")
	go func() {
		// This we do in case l.Stop() doesn't work.
		// It's a way to force quit.
		<-c
		os.Exit(1)
	}()
	l.Stop()
}
