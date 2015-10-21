// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type Measurement struct {
	UnixTime    int64
	Temperature float32
	Humidity    float32
}

func (x *Measurement) Update(lag float32, m Measurement) {
	x.UnixTime = m.UnixTime
	x.Temperature = (1-lag)*x.Temperature + lag*m.Temperature
	x.Humidity = (1-lag)*x.Humidity + lag*m.Humidity
}

func (x Measurement) String() string {
	t := time.Unix(x.UnixTime, 0).Format(time.Stamp)
	return fmt.Sprintf("%v: %.1f \u2103C at %.1f%% humidity", t, x.Temperature, x.Humidity)
}

func (x Measurement) CSV() string {
	t := time.Unix(x.UnixTime, 0).Format("2006-01-02,15:04:05")
	return fmt.Sprintf("%v,%.2f,%.2f", t, x.Temperature, x.Humidity)
}

// One days worth of measurements should take up about 1,382,400 bytes.
// This means we should have no problem storing a week of measurements.
// After that, we simplify the measurements, so we have once every minute.
// And after that, we can simplify further, averaging the hours. Eventually,
// we'll resort to using a database instead of this here. Probably should
// do that from the start.
type Series []Measurement

func (s *Series) Add(m Measurement) {
	*s = append(*s, m)
}

type Monitor struct {
	Belief Measurement
	Series Series
	Writer *bufio.Writer

	lag  float32
	file *os.File
}

func NewMonitor(file string, lag float32) (*Monitor, error) {
	m := &Monitor{
		lag: lag,
	}

	if file != "" {
		var err error
		m.file, err = os.Create(file)
		if err != nil {
			return nil, err
		}

		m.Writer = bufio.NewWriter(m.file)
		m.Writer.WriteString("time,temperature,humidity\n")
	}
	return m, nil
}

func (m *Monitor) Update(x Measurement) {
	m.Belief.Update(m.lag, x)
	m.Series.Add(x)
	if m.Writer != nil {
		m.Writer.WriteString(x.CSV())
		m.Writer.WriteRune('\n')
	}
}

func (m *Monitor) Close() {
	if m.Writer != nil {
		m.Writer.Flush()
		m.file.Close()
	}
}
