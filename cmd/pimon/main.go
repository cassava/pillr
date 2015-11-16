// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/cassava/pillr/guitar"
	"github.com/cassava/pillr/led"
	"github.com/goulash/xdg"
	"github.com/spf13/cobra"
)

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

		Run()
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
	pf.IntVarP(&Conf.PinSensor, "pin-warning", "S", Conf.PinSensor, "pin number for sensor")

	pimonCmd.AddCommand(VersionCmd)
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

func exitIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
