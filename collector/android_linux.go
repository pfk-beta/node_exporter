// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !noandroid && (openbsd || linux || darwin)
// +build !noandroid
// +build openbsd linux darwin

package collector

import (
	"os/exec"
	"regexp"
	"strconv"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	androidLabelNames = []string{}
)

type androidCollector struct {
	batteryChargingDesc *prometheus.Desc
	batteryLevelDesc    *prometheus.Desc
	batteryTempDesc     *prometheus.Desc
	logger              log.Logger
}

func init() {
	registerCollector("android", true, NewAndroidCollector)
}

// NewAndroidCollector returns a new Collector exposing dumpsys and getprop values.
func NewAndroidCollector(logger log.Logger) (Collector, error) {

	subsystem := "android"

	batteryChargingDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "battery_charging"),
		"Charging state.",
		androidLabelNames, nil,
	)

	batteryLevelDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "battery_level"),
		"Level of charged battery.",
		androidLabelNames, nil,
	)

	batteryTempDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "battery_temperature"),
		"Temperature of battery.",
		androidLabelNames, nil,
	)

	return &androidCollector{
		batteryChargingDesc: batteryChargingDesc,
		batteryLevelDesc:    batteryLevelDesc,
		batteryTempDesc:     batteryTempDesc,
		logger:              logger,
	}, nil
}

func (c *androidCollector) Update(ch chan<- prometheus.Metric) error {
	charging, level, temp := c.GetStats()

	ch <- prometheus.MustNewConstMetric(
		c.batteryChargingDesc, prometheus.GaugeValue,
		charging,
	)
	ch <- prometheus.MustNewConstMetric(
		c.batteryLevelDesc, prometheus.GaugeValue,
		level,
	)
	ch <- prometheus.MustNewConstMetric(
		c.batteryTempDesc, prometheus.GaugeValue,
		temp,
	)
	return nil
}

// GetStats returns filesystem stats.
func (c *androidCollector) GetStats() (float64, float64, float64) {
	// get dumpsys for battery
	out, _ := exec.Command("/system/bin/dumpsys", "battery").Output()
	output := string(out[:])

	batteryChargingRegex := regexp.MustCompile("AC powered: (?P<charging>\\w+)")
	batteryLevelRegex := regexp.MustCompile("level: (?P<level>\\d+)")
	batteryTempRegex := regexp.MustCompile("temperature: (?P<temp>\\d+)")

	charging_str := batteryChargingRegex.FindStringSubmatch(output)[1]

	charging_value := 0.0
	if charging_str == "true" {
		charging_value = 1.0
	}

	level_value, _ := strconv.Atoi(batteryLevelRegex.FindStringSubmatch(output)[1])
	temp_value, _ := strconv.Atoi(batteryTempRegex.FindStringSubmatch(output)[1])

	return charging_value, float64(level_value), float64(temp_value) / 10
}
