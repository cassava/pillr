package led

import "time"

var (
	Heartbeat500  = []time.Duration{100 * time.Millisecond, 500 * time.Millisecond}
	Heartbeat1000 = []time.Duration{100 * time.Millisecond, 1000 * time.Millisecond}
	Heartbeat5000 = []time.Duration{100 * time.Millisecond, 5000 * time.Millisecond}

	FastBlink = []time.Duration{100 * time.Millisecond}

	Elevated = []time.Duration{100 * time.Millisecond, 1000 * time.Millisecond}
	High     = []time.Duration{50 * time.Millisecond, 500 * time.Millisecond}
	Severe   = []time.Duration{50 * time.Millisecond, 250 * time.Millisecond}
	Extreme  = []time.Duration{50 * time.Millisecond, 50 * time.Millisecond}
)
