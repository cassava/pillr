// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/cassava/pillr/guitar"
	"github.com/cassava/pillr/led"
	"github.com/goulash/xdg"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/spf13/cobra"
)

var Conf = &Configuration{
	Listen:   ":8080",
	Conserve: false,
	Interval: 10 * time.Second,

	Patterns: PatternConfiguration{
		Low:      []time.Duration{0},
		Moderate: led.Moderate,
		Elevated: led.Elevated,
		High:     led.High,
		Severe:   led.Severe,
		Extreme:  led.Extreme,
	},
}

// Configuration type {{{

const (
	configEnv      = "PIMON_CONFIG"
	configSuffix   = "pimon/conf.toml"
	databaseSuffix = "pimon/measurements.dat"
	lockSuffix     = "pimon/lock.pid"
)

type Configuration struct {
	// Listen defines what port we should listen to for online access.
	// If left empty, online access is disabled.
	Listen string `toml:"listen"`

	// Conserve defines if we only store entries that differ from previous entries.
	Conserve bool `toml:"conserve"`

	// Interval defines the minimum time between measurements.
	Interval time.Duration `toml:"interval"`

	PinWarningLED   int `toml:"pin_warning_led"`
	PinHeartbeatLED int `toml:"pin_heartbeat_led"`
	PinSensor       int `toml:"pin_sensor"`

	Patterns PatternConfiguration `toml:"patterns"`
}

type PatternConfiguration struct {
	Low      []time.Duration `toml:"low"`
	Moderate []time.Duration `toml:"moderate"`
	Elevated []time.Duration `toml:"elevated"`
	High     []time.Duration `toml:"high"`
	Severe   []time.Duration `toml:"severe"`
	Extreme  []time.Duration `toml:"extreme"`
}

func (pc PatternConfiguration) Get(d guitar.Danger) []time.Duration {
	switch d {
	case guitar.Low:
		return pc.Low
	case guitar.Moderate:
		return pc.Moderate
	case guitar.Elevated:
		return pc.Elevated
	case guitar.High:
		return pc.High
	case guitar.Severe:
		return pc.Severe
	case guitar.Extreme:
		return pc.Extreme
	default:
		log.Error("Unknown danger level received.")
		return nil
	}
}

func (c Configuration) Assert() {
	if c.PinWarningLED <= 0 {
		log.Fatal("warning LED pin unspecified")
	}
	if c.PinSensor <= 0 {
		log.Fatal("sensor pin unspecified")
	}
	if c.Interval < 0 {
		log.Fatal("measurment interval is invalid")
	}
}

// }}}

// Main (pimon) command {{{

var (
	database string
	conf     string
)

var pimonCmd = &cobra.Command{
	Use:   "pimon",
	Short: "monitor temperature and humidity",
	Long: `Pimon monitors the temperature and humidity and warns you
if if is not in the safe range defined by Larrivee.

  If pimon is run with default options and without any specific command,
  it will read all the configuration files it finds in the XDG config path.
  It will also store any measurements in XDG_DATA_HOME/pimon/th.csv.
`,
	Run: func(cmd *cobra.Command, args []string) {
		Conf.Assert()

		exitIf(pimonLock())
		defer pimonUnlock()

		exitIf(embd.InitGPIO())
		defer embd.CloseGPIO()

		done := make(chan struct{})
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		nl := &WarningLED{led.New(Conf.PinWarningLED), guitar.Low}
		csv, err := NewCSVPersister(xdg.UserData(databaseSuffix))
		if err != nil {
			log.Error("error persisting: ", err)
			return
		}
		m, err := NewMonitor(csv, 0.1)
		if err != nil {
			log.Error("error reading persistent file: ", err)
			return
		}
		g := guitar.Larrivee

		go Serve(Conf.Listen, m)
		go WatchSensor(Conf.PinSensor, done, func(x Measurement) {
			m.Update(x)
			d := g.Threat(x.Humidity)
			nl.Update(d)
			log.WithFields(log.Fields{
				"danger": d.String(),
			}).Info(x)
		})

		<-c
		close(done)
		m.Close()
		nl.LED.Stop()
	},
}

func pimonInit() {
	pf := pimonCmd.PersistentFlags()
	pf.StringVar(&database, "database", "", "read from and store measurements in this file")
	pf.StringVar(&Conf.Listen, "listen", Conf.Listen, "enable online access at this port")
	pf.BoolVarP(&Conf.Conserve, "conserve", "c", Conf.Conserve, "only store differing entries")
	pf.DurationVarP(&Conf.Interval, "interval", "i", Conf.Interval, "minimum time between measurements")
	pf.IntVarP(&Conf.PinWarningLED, "pin-warning", "W", Conf.PinWarningLED, "pin number for warning LED")
	pf.IntVarP(&Conf.PinHeartbeatLED, "pin-heartbeat", "H", Conf.PinHeartbeatLED, "pin number for system LED")
	pf.IntVarP(&Conf.PinSensor, "pin-sensor", "S", Conf.PinSensor, "pin number for sensor")

	pimonCmd.AddCommand(versionCmd)
}

func mergeConf() {
	merge := func(file string) {
		_, err := toml.DecodeFile(file, &Conf)
		if err != nil {
			log.Errorf("error reading configuration file %v: %v", file, err)
		}
	}

	cp := os.Getenv(configEnv)
	if cp != "" {
		merge(cp)
		return
	}
	xdg.MergeConfigR(configSuffix, func(file string) error { merge(file); return nil })
}

func pimonLock() error {
	if xdg.FindRuntime(lockSuffix) != "" {
		return fmt.Errorf("pimon lock file already exists: %s", xdg.UserRuntime(lockSuffix))
	}

	f, err := xdg.OpenRuntime(lockSuffix, os.O_EXCL|os.O_CREATE|os.O_WRONLY)
	if err != nil {
		return fmt.Errorf("cannot create pimon lock file: %v", err)
	}
	fmt.Fprintln(f, os.Getpid())
	f.Close()
	return nil
}

func pimonUnlock() {
	_ = os.Remove(xdg.UserRuntime(lockSuffix))
}

func main() {
	mergeConf()
	pimonInit()
	pimonCmd.Execute()
}

// }}}

// Version command {{{

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version and date information",
	Long:  "Show the official version number of pimon, as well as the release date.",
	Run: func(cmd *cobra.Command, args []string) {
		writeVersion(os.Stdout)
	},
}

type programInfo struct {
	Name      string
	Author    string
	Email     string
	Version   string
	Date      string
	Homepage  string
	Copyright string
	License   string
}

var versionTmpl = template.Must(template.New("version").Parse(
	`{{.Name}} version {{.Version}} ({{.Date}})
Copyright {{.Copyright}}, {{.Author}} <{{.Email}}>

You may find {{.Name}} on the Internet at

    {{.Homepage}}

Please report any bugs you may encounter.

The source code of {{.Name}} is licensed under the {{.License}} license.
`))

var progInfo = programInfo{
	Name:      "pimon",
	Author:    "Ben Morgan",
	Email:     "neembi@gmail.com",
	Version:   "0.1",
	Date:      time.Now().Format("2 January 2006"),
	Copyright: time.Now().Format("2006"),
	Homepage:  "https://github.com/cassava/pillr",
	License:   "MIT",
}

func writeVersion(w io.Writer) {
	versionTmpl.Execute(w, progInfo)
}

// }}}

// Utility functions {{{

func exitIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// }}}
