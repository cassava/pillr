// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cassava/pillr/guitar"
	"github.com/cassava/pillr/led"
	"github.com/d2r2/go-dht"
	"github.com/goulash/xdg"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

func listen(pin int, ch chan<- Measurement, done <-chan struct{}) {
	read := func() (m Measurement) {
		before := time.Now()
		after := time.Now()
		for after.Sub(before) < Conf.Interval {
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

type WarningLED struct {
	LED    *led.LED
	Threat guitar.Danger
}

func (wl *WarningLED) Update(d guitar.Danger) {
	p := Conf.Patterns.Get(d)
	if len(p) < 2 {
		wl.LED.Stop()
		return
	}

	wl.LED.Blink(p...)
}

func Run() {
	exitIf(embd.InitGPIO())
	defer embd.CloseGPIO()

	ch := make(chan Measurement, 1)
	done := make(chan struct{})
	go listen(Conf.PinSensor, ch, done)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	nl := &WarningLED{led.New(Conf.PinWarningLED), guitar.Low}
	csv, err := NewCSVPersister(xdg.UserData(databaseSuffix))
	if err != nil {
		log.Fatal(err)
	}
	m, _ := NewMonitor(csv, 0.1)
	g := guitar.Larrivee

	go Serve(Conf.Listen, m)

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
