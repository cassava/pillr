// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package led

import "time"

var (
	Heartbeat500  = []time.Duration{100 * time.Millisecond, 500 * time.Millisecond}
	Heartbeat1000 = []time.Duration{100 * time.Millisecond, 1000 * time.Millisecond}
	Heartbeat5000 = []time.Duration{100 * time.Millisecond, 5000 * time.Millisecond}

	FastBlink = []time.Duration{100 * time.Millisecond}

	Moderate = []time.Duration{500 * time.Millisecond, 5000 * time.Millisecond}
	Elevated = []time.Duration{100 * time.Millisecond, 1000 * time.Millisecond}
	High     = []time.Duration{50 * time.Millisecond, 500 * time.Millisecond}
	Severe   = []time.Duration{50 * time.Millisecond, 250 * time.Millisecond}
	Extreme  = []time.Duration{50 * time.Millisecond, 50 * time.Millisecond}
)
