// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"os"
)

type Persister interface {
	Persist(m Measurement) error
	Close() error
}

type csvPersister struct {
	file *os.File
	buf  *bufio.Writer
	w    *csv.Writer
}

func NewCSVPersister(path string) (*csvPersister, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewWriter(file)
	w := csv.NewWriter(buf)
	return &csvPersister{file, buf, w}, nil
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
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewWriter(file)
	enc := gob.NewEncoder(buf)
	return &gobPersister{file, buf, enc}, nil
}
