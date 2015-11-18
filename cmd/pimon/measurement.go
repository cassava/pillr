// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const measurementTimeFormat = "2006-01-02 15:04:05"

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

func (x Measurement) Record() []string {
	return []string{time.Unix(x.UnixTime, 0).Format(measurementTimeFormat),
		strconv.FormatFloat(float64(x.Temperature), 'f', 1, 32),
		strconv.FormatFloat(float64(x.Humidity), 'f', 1, 32),
	}
}

func (x Measurement) MarshalJSON() ([]byte, error) {
	r := x.Record()
	return []byte(fmt.Sprintf(`{"time": "%s", "temperature": %s, "humidity": %s}`, r[0], r[1], r[2])), nil
}

func (m *Measurement) FromRecord(rs []string) error {
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

func (x Measurement) Marshal(v url.Values) string {
	t := v.Get("type")
	switch t {
	case "csv":
		return strings.Join(x.Record(), ",")
	case "string":
		return x.String()
	case "tabular":
		r := x.Record()
		return fmt.Sprintf("time:        %s\ntemperature: %s\nhumidity:    %s", r[0], r[1], r[2])
	case "json":
		fallthrough
	default:
		s, _ := x.MarshalJSON()
		return string(s)
	}
}

func (x Measurement) Same(y Measurement) bool {
	return x.Temperature == y.Temperature && x.Humidity == y.Humidity
}

// One days worth of measurements should take up about 1,382,400 bytes.
// This means we should have no problem storing a week of measurements.
// After that, we simplify the measurements, so we have once every minute.
// And after that, we can simplify further, averaging the hours. Eventually,
// we'll resort to using a database instead of this here. Probably should
// do that from the start.
type Series []Measurement

func (s *Series) Add(m Measurement) { *s = append(*s, m) }
func (s Series) Len() int           { return len(s) }
func (s Series) Top() Measurement   { return s[len(s)-1] }

func (s Series) Marshal(w io.Writer, v url.Values) {
	t := v.Get("type")
	switch t {
	case "csv":
		cw := csv.NewWriter(w)
		for _, x := range s {
			err := cw.Write(x.Record())
			if err != nil {
				log.Errorln("marshalling series to csv: ", err)
				break
			}
		}
		cw.Flush()
	case "json":
		fallthrough
	default:
		jw := json.NewEncoder(w)
		jw.Encode(s)
	}
}
