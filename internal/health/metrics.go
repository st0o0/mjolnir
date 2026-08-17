package health

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	upsStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_status",
		Help: "UPS status: 1=online, 0=on-battery, -1=unknown",
	}, []string{"ups"})

	batteryCharge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_battery_charge_percent",
		Help: "Battery charge percentage",
	}, []string{"ups"})

	loadPercent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_load_percent",
		Help: "UPS load percentage",
	}, []string{"ups"})

	inputVoltage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_input_voltage",
		Help: "Input voltage",
	}, []string{"ups"})

	outputVoltage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_output_voltage",
		Help: "Output voltage",
	}, []string{"ups"})

	batteryVoltage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_battery_voltage",
		Help: "Battery voltage",
	}, []string{"ups"})
)

func UpdateMetrics(upsName string, vars map[string]string) {
	if s, ok := vars["ups.status"]; ok {
		upsStatus.WithLabelValues(upsName).Set(statusToFloat(s))
	}

	setGaugeFromVar(batteryCharge, upsName, vars, "battery.charge")
	setGaugeFromVar(loadPercent, upsName, vars, "ups.load")
	setGaugeFromVar(inputVoltage, upsName, vars, "input.voltage")
	setGaugeFromVar(outputVoltage, upsName, vars, "output.voltage")
	setGaugeFromVar(batteryVoltage, upsName, vars, "battery.voltage")
}

func setGaugeFromVar(gauge *prometheus.GaugeVec, upsName string, vars map[string]string, key string) {
	if val, ok := vars[key]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			gauge.WithLabelValues(upsName).Set(f)
		}
	}
}

func statusToFloat(s string) float64 {
	switch {
	case strings.Contains(s, "OL"):
		return 1
	case strings.Contains(s, "OB"):
		return 0
	default:
		return -1
	}
}
