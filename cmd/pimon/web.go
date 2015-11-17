// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

var monitor *Monitor

func serveBelief(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, monitor.Belief().Marshal(r.URL.Query()))
}

func serveLatest(w http.ResponseWriter, r *http.Request) {
	s := monitor.Series()
	if len(s) == 0 {
		http.Error(w, "no measurement data", 500)
		return
	}
	fmt.Fprintln(w, s.Top().Marshal(r.URL.Query()))
}

func init() {
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
