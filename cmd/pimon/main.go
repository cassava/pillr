// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cassava/pillr/guitar"
	"github.com/cassava/pillr/led"
	"github.com/d2r2/go-dht"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

var (
	// If conserve is true, we should (can) set interval to 0.
	conserve = flag.Bool("conserve", true, "only store entries that differ from the previous")
	pinnr    = flag.Int("led", -1, "gpio pin number of LED")
	dhtnr    = flag.Int("dht", -1, "gpio pin number of DHT22 sensor")
	mondb    = flag.String("output", "pimon.csv", "file to write data to")
	interval = flag.Duration("interval", 2*time.Second, "minimum time between measurements")
)

func listen(pin int, ch chan<- Measurement, done <-chan struct{}) {
	read := func() (m Measurement) {
		before := time.Now()
		after := before
		for after.Sub(before) <= *interval {
			t, h, r, err := dht.ReadDHTxxWithRetry(dht.DHT22, pin, false, 10)
			if err != nil {
				log.WithFields(log.Fields{"retries": r}).Errorf("DHT22: %s", err)
				continue
			}
			log.WithFields(log.Fields{"retries": r}).Debugf("DHT22: temp=%v humidity=%v", t, h)

			after = time.Now()
			m = Measurement{after.Unix(), t, h}
		}
		return
	}

outer:
	for {
		select {
		case <-done:
			break outer
		case ch <- read():
			continue
		}
	}

	close(ch)
}

type NotifyLED struct {
	LED    *led.LED
	Threat guitar.Danger
}

func (nl *NotifyLED) Update(d guitar.Danger) {
	if nl.Threat == d {
		return
	}
	nl.Threat = d

	switch d {
	case guitar.Low:
		nl.LED.Stop()
	case guitar.Moderate:
		nl.LED.Blink(led.Moderate...)
	case guitar.Elevated:
		nl.LED.Blink(led.Elevated...)
	case guitar.High:
		nl.LED.Blink(led.High...)
	case guitar.Severe:
		nl.LED.Blink(led.Severe...)
	case guitar.Extreme:
		nl.LED.Blink(led.Extreme...)
	default:
		log.Error("Unknown danger level received.")
	}
}

func main() {
	flag.Parse()

	if *pinnr < 0 {
		log.Fatal("LED pin unspecified")
	}
	if *dhtnr < 0 {
		log.Fatal("DHT22 pin unspecified")
	}
	if *interval < 0 {
		*interval = 0
	}

	exitIf(embd.InitGPIO())
	defer embd.CloseGPIO()

	ch := make(chan Measurement, 1)
	done := make(chan struct{})
	go listen(*dhtnr, ch, done)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	nl := &NotifyLED{led.New(*pinnr), guitar.Low}
	m, err := NewMonitor(*mondb, 0.1)
	if err != nil {
		log.Fatal(err)
	}
	g := guitar.Larrivee
outer:
	for {
		select {
		case <-c:
			log.Info("Exiting...")
			break outer
		case x := <-ch:
			m.Update(x)
			d := g.Threat(x.Humidity)
			nl.Update(d)
			log.WithFields(log.Fields{
				"danger": d.String(),
			}).Info(x)
		}
	}

	// Signal listener that it should close.
	close(done)
	m.Close()

	go func() {
		// This we do in case l.Stop() doesn't work.
		// It's a way to force quit.
		<-c
		log.Error("Forcing exit.")
		os.Exit(1)
	}()
	nl.LED.Stop()
}

func exitIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
