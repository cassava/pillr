// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cassava/pillr/guitar"
	"github.com/cassava/pillr/led"
	"github.com/d2r2/go-dht"
)

type WarningLED struct {
	LED    *led.LED
	Threat guitar.Danger
}

func (wl *WarningLED) Update(d guitar.Danger) {
	if wl.Threat == d {
		return
	}

	wl.Threat = d
	p := Conf.Patterns.Get(d)
	if len(p) < 2 {
		wl.LED.Stop()
		return
	}

	wl.LED.Blink(p...)
}

func WatchSensor(pin int, done <-chan struct{}, f func(Measurement)) {
	ch := make(chan Measurement, 1)

	read := func() {
		before := time.Now()
		after := time.Now()
		for after.Sub(before) <= Conf.Interval {
			t, h, r, err := dht.ReadDHTxxWithRetry(dht.DHT22, pin, false, 10)
			if err != nil {
				log.WithFields(log.Fields{"retries": r}).Errorf("DHT22: %s", err)
				continue
			}
			log.WithFields(log.Fields{"retries": r}).Debugf("DHT22: temp=%v humidity=%v", t, h)

			after = time.Now()
			ch <- Measurement{after.Unix(), t, h}
		}
	}

	go read()
	for {
		select {
		case <-done:
			return
		case x := <-ch:
			f(x)
		}
	}
}
