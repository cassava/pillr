// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

type Measurement struct {
	UnixTime    int64
	Temperature float32
	Humidity    float32
}

// Measurement implemenation {{{

const measurementTimeFormat = "2006-01-02 15:04:05"

func (x *Measurement) Update(lag float32, m Measurement) {
	x.UnixTime = m.UnixTime
	x.Temperature = (1-lag)*x.Temperature + lag*m.Temperature
	x.Humidity = (1-lag)*x.Humidity + lag*m.Humidity
}

func (x Measurement) String() string {
	t := time.Unix(x.UnixTime, 0).Format(time.Stamp)
	return fmt.Sprintf("%v: %.1f C at %.1f%% humidity", t, x.Temperature, x.Humidity)
}

func (x Measurement) MarshalRecord() []string {
	return []string{time.Unix(x.UnixTime, 0).Format(measurementTimeFormat),
		strconv.FormatFloat(float64(x.Temperature), 'f', 1, 32),
		strconv.FormatFloat(float64(x.Humidity), 'f', 1, 32),
	}
}

func (x Measurement) MarshalCSV() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("time,temperature,humidity\n")
	r := x.MarshalRecord()
	buf.WriteString(r[0])
	buf.WriteRune(',')
	buf.WriteString(r[1])
	buf.WriteRune(',')
	buf.WriteString(r[2])
	buf.WriteRune('\n')
	return buf.Bytes(), nil
}

func (x Measurement) MarshalJSON() ([]byte, error) {
	r := x.MarshalRecord()
	return []byte(fmt.Sprintf(`{"time": "%s", "temperature": %s, "humidity": %s}`, r[0], r[1], r[2])), nil
}

func (m *Measurement) UnmarshalRecord(rs []string) error {
	if len(rs) != 3 {
		return errors.New("invalid record length")
	}

	t, err := time.Parse(measurementTimeFormat, rs[0])
	if err != nil {
		return err
	}
	m.UnixTime = t.Unix()
	f, err := strconv.ParseFloat(rs[1], 32)
	if err != nil {
		return err
	}
	m.Temperature = float32(f)
	f, err = strconv.ParseFloat(rs[2], 32)
	if err != nil {
		return err
	}
	m.Humidity = float32(f)
	return nil
}

// Same returns true when the measurement itself is the same.
func (x Measurement) Same(y Measurement) bool {
	return x.Temperature == y.Temperature && x.Humidity == y.Humidity
}

// }}}

// One days worth of measurements should take up about 1,382,400 bytes.
// This means we should have no problem storing a week of measurements.
// After that, we simplify the measurements, so we have once every minute.
// And after that, we can simplify further, averaging the hours. Eventually,
// we'll resort to using a database instead of this here. Probably should
// do that from the start.
type Series []Measurement

// Series implementation {{{

func (s *Series) Add(m Measurement) { *s = append(*s, m) }
func (s Series) Len() int           { return len(s) }
func (s Series) Top() Measurement   { return s[len(s)-1] }

func (s Series) MarshalCSV() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("time,temperature,humidity\n")
	for _, x := range s {
		r := x.MarshalRecord()
		buf.WriteString(r[0])
		buf.WriteRune(',')
		buf.WriteString(r[1])
		buf.WriteRune(',')
		buf.WriteString(r[2])
		buf.WriteRune('\n')
	}
	return buf.Bytes(), nil
}

func (s *Series) UnmarshalCSV(bs []byte) error {
	buf := bytes.NewBuffer(bs)
	cr := csv.NewReader(buf)
	for {
		r, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		var x Measurement
		err = x.UnmarshalRecord(r)
		if err != nil {
			return err
		}

		*s = append(*s, x)
	}
	return nil
}

// }}}
