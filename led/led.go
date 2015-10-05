package led

import (
	"fmt"
	"os"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

type LED struct {
	pin   embd.DigitalPin
	state bool

	blink bool
	ok    chan struct{}
	stop  chan struct{}
}

func New(pin int) *LED {
	p, err := embd.NewDigitalPin(pin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error instantiating pin %v: %s\n", pin, err)
		os.Exit(1)
	}
	p.SetDirection(embd.Out)
	return &LED{pin: p}
}

func (l *LED) On() {
	l.pin.Write(embd.High)
	l.state = true
}

func (l *LED) Off() {
	l.pin.Write(embd.Low)
	l.state = false
}

func (l *LED) State() bool {
	return l.state
}

func (l *LED) Toggle() {
	if l.state {
		l.Off()
	} else {
		l.On()
	}
}

func (l *LED) Sustain(t time.Duration) {
	l.On()
	time.Sleep(t)
	l.Off()
}

func (l *LED) Blink(pattern ...time.Duration) {
	if len(pattern) == 0 {
		return
	}

	l.Stop()
	go func() {
		l.blink = true
		l.stop = make(chan struct{})
		l.ok = make(chan struct{})

		l.On()
	outer:
		for {
			for _, t := range pattern {
				select {
				case <-time.After(t):
					l.Toggle()
				case <-l.stop:
					break outer
				}
			}
		}

		l.Off()
		close(l.ok)
	}()
}

func (l *LED) Stop() {
	if l.blink {
		close(l.stop)
		<-l.ok
		l.blink = false
	}
	l.Off()
}
