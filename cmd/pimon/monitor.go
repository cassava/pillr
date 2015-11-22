// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"io/ioutil"
	"os"
	"sync"
)

type Monitor struct {
	sync.RWMutex

	belief Measurement
	series Series
	p      Persister
	lag    float32
}

// Monitor implementation {{{

func NewMonitor(p Persister, lag float32) (*Monitor, error) {
	s, err := p.ReadAll()
	if err != nil {
		return nil, err
	}
	return &Monitor{
		lag:    lag,
		p:      p,
		series: s,
	}, nil
}

func (m *Monitor) Belief() Measurement {
	m.RLock()
	defer m.RUnlock()
	return m.belief
}

func (m *Monitor) Series() Series {
	m.RLock()
	defer m.RUnlock()
	return m.series
}

func (m *Monitor) Update(x Measurement) {
	m.Lock()
	defer m.Unlock()

	m.belief.Update(m.lag, x)
	if Conf.Conserve && m.series.Len() != 0 && m.series.Top().Same(x) {
		return
	}

	m.series.Add(x)
	if m.p != nil {
		m.p.Persist(x)
	}
}

func (m *Monitor) Close() {
	if m.p != nil {
		m.Lock()
		defer m.Unlock()
		m.p.Close()
	}
}

// }}}

type Persister interface {
	ReadAll() (Series, error)
	Persist(m Measurement) error
	Close() error
}

// Persister implementation {{{
// CSV Persister {{{

type csvPersister struct {
	file *os.File
	w    *csv.Writer
}

func NewCSVPersister(path string) (*csvPersister, error) {
	file, err := open(path)
	if err != nil {
		return nil, err
	}

	w := csv.NewWriter(file)
	return &csvPersister{file, w}, nil
}

func (p *csvPersister) ReadAll() (Series, error) {
	_, err := p.file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	defer p.file.Seek(0, 2)
	bs, err := ioutil.ReadAll(p.file)
	if err != nil {
		return nil, err
	}
	var s Series
	err = s.UnmarshalCSV(bs)
	return s, err
}

func (p *csvPersister) Persist(m Measurement) error {
	return p.w.Write(m.MarshalRecord())
}

func (p *csvPersister) Close() error {
	p.w.Flush()
	return p.file.Close()
}

// }}}

// Gob Persister {{{

type gobPersister struct {
	file *os.File
	buf  *bufio.Writer
	enc  *gob.Encoder
}

func (p *gobPersister) Persist(m Measurement) error {
	return p.enc.Encode(m.MarshalRecord())
}

func (p *gobPersister) Close() error {
	p.buf.Flush()
	return p.file.Close()
}

func NewGobPersister(path string) (*gobPersister, error) {
	file, err := open(path)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewWriter(file)
	enc := gob.NewEncoder(buf)
	return &gobPersister{file, buf, enc}, nil
}

// }}}

func open(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
}

// }}}
