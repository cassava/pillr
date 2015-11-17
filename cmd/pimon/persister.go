// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"io"
	"os"
)

type Persister interface {
	ReadAll() (Series, error)
	Persist(m Measurement) error
	Close() error
}

func open(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
}

type csvPersister struct {
	file *os.File
	buf  *bufio.Writer
	w    *csv.Writer
}

func NewCSVPersister(path string) (*csvPersister, error) {
	file, err := open(path)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewWriter(file)
	w := csv.NewWriter(buf)
	return &csvPersister{file, buf, w}, nil
}

func (p *csvPersister) ReadAll() (Series, error) {
	_, err := p.file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	defer p.file.Seek(0, 2)
	r := csv.NewReader(p.file)
	var s Series
	for {
		var rs []string
		rs, err = r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		var x Measurement
		err = x.FromRecord(rs)
		if err != nil {
			return nil, err
		}
		s.Add(x)
	}
	return s, nil
}

func (p *csvPersister) Persist(m Measurement) error {
	return p.w.Write(m.Record())
}

func (p *csvPersister) Close() error {
	p.buf.Flush()
	return p.file.Close()
}

type gobPersister struct {
	file *os.File
	buf  *bufio.Writer
	enc  *gob.Encoder
}

func (p *gobPersister) Persist(m Measurement) error {
	return p.enc.Encode(m.Record())
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
