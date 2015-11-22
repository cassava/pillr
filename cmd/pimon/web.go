// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

var monitor *Monitor

func init() {
	http.HandleFunc("/series", serveSeries)
	http.HandleFunc("/belief", serveBelief)
	http.HandleFunc("/latest", serveLatest)
}

func Serve(listen string, m *Monitor) {
	monitor = m
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Errorln(err)
	}
}

func serveSeries(w http.ResponseWriter, r *http.Request) { serveStruct(w, r, monitor.Series()) }
func serveBelief(w http.ResponseWriter, r *http.Request) { serveStruct(w, r, monitor.Belief()) }
func serveLatest(w http.ResponseWriter, r *http.Request) {
	s := monitor.Series()
	if len(s) == 0 {
		http.Error(w, "no measurement data", 500)
		return
	}
	serveStruct(w, r, s.Top())
}

type csvMarshaler interface {
	MarshalCSV() ([]byte, error)
}

func serveStruct(w http.ResponseWriter, r *http.Request, v interface{}) {
	q := r.URL.Query()
	switch q.Get("type") {
	case "string":
		t, ok := v.(fmt.Stringer)
		if !ok {
			http.Error(w, "this endpoint does not support string type", 400)
			return
		}
		fmt.Fprintln(w, t)
	case "csv":
		t, ok := v.(csvMarshaler)
		if !ok {
			http.Error(w, "this endpoint does not support csv type", 400)
			return
		}
		bs, err := t.MarshalCSV()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(bs)
	case "", "json": // this is the default if type is not defined
		bs, err := json.Marshal(v)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(bs)
	default:
		http.Error(w, "type unknown", 400)
	}
}
