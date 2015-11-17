// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "sync"

type Monitor struct {
	*sync.RWMutex

	belief Measurement
	series Series
	p      Persister
	lag    float32
}

func NewMonitor(p Persister, lag float32) (*Monitor, error) {
	s, err := p.ReadAll()
	if err != nil {
		return nil, err
	}
	return &Monitor{
		RWMutex: &sync.RWMutex{},
		lag:     lag,
		p:       p,
		series:  s,
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
