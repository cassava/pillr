// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package guitar

type Danger int

const (
	Low      Danger = iota // No danger
	Moderate               // Notify after 1 week
	Elevated               // Notify after 3 days
	High                   // Notify after 6 hours
	Severe                 // Notify after 15 minutes
	Extreme                // Notify after 3 minutes
)

type Levels struct {
	Gradient []float64
	Risk     []Danger
	Details  *Notification
}

type Notification struct {
	SubjectTmpl  string
	BodyTmpl     string
	ShortEffects []string
	LongEffects  []string
}

func (l Levels) Threat(v float64) Danger {
	for i, g := range l.Gradient {
		if v < g {
			return l.Risk[i]
		}
	}
	return Extreme
}
