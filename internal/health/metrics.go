package health

import (
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var statusFlags = []string{
	"OL", "OB", "LB", "HB", "RB",
	"CHRG", "DISCHRG", "BYPASS", "CAL", "OFF",
	"OVER", "TRIM", "BOOST", "FSD",
}

var deviceInfoKeys = []struct {
	nutVar string
	label  string
}{
	{"device.model", "model"},
	{"device.mfr", "mfr"},
	{"device.serial", "serial"},
	{"device.type", "type"},
}

type MetricWriter struct {
	mu     sync.Mutex
	gauges map[string]*prometheus.GaugeVec
	reg    prometheus.Registerer

	statusGauge *prometheus.GaugeVec
	deviceInfo  *prometheus.GaugeVec
	scrapeError *prometheus.GaugeVec
}

func NewMetricWriter(reg prometheus.Registerer) *MetricWriter {
	statusGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_status",
		Help: "UPS status flags: 1=active, 0=inactive",
	}, []string{"ups", "flag"})

	deviceInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_device_info",
		Help: "UPS device information",
	}, []string{"ups", "model", "mfr", "serial", "type"})

	scrapeError := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_scrape_error",
		Help: "1 if the last scrape for this UPS failed",
	}, []string{"ups"})

	reg.MustRegister(statusGauge, deviceInfo, scrapeError)

	return &MetricWriter{
		gauges:      make(map[string]*prometheus.GaugeVec),
		reg:         reg,
		statusGauge: statusGauge,
		deviceInfo:  deviceInfo,
		scrapeError: scrapeError,
	}
}

func (w *MetricWriter) Write(upsName string, vars map[string]string) {
	w.scrapeError.WithLabelValues(upsName).Set(0)
	w.writeStatus(upsName, vars)
	w.writeDeviceInfo(upsName, vars)
	w.writeDynamicGauges(upsName, vars)
}

func (w *MetricWriter) SetError(upsName string) {
	w.scrapeError.WithLabelValues(upsName).Set(1)
}

func (w *MetricWriter) writeStatus(upsName string, vars map[string]string) {
	status := vars["ups.status"]
	activeFlags := make(map[string]bool)
	for _, f := range strings.Fields(status) {
		activeFlags[f] = true
	}
	for _, flag := range statusFlags {
		val := 0.0
		if activeFlags[flag] {
			val = 1.0
		}
		w.statusGauge.WithLabelValues(upsName, flag).Set(val)
	}
}

func (w *MetricWriter) writeDeviceInfo(upsName string, vars map[string]string) {
	labels := make([]string, len(deviceInfoKeys)+1)
	labels[0] = upsName
	for i, dk := range deviceInfoKeys {
		labels[i+1] = vars[dk.nutVar]
	}
	w.deviceInfo.WithLabelValues(labels...).Set(1)
}

func (w *MetricWriter) writeDynamicGauges(upsName string, vars map[string]string) {
	for nutVar, val := range vars {
		if nutVar == "ups.status" {
			continue
		}
		if strings.HasPrefix(nutVar, "device.") || strings.HasPrefix(nutVar, "driver.") {
			continue
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			continue
		}
		metricName := varToMetricName(nutVar)
		gauge := w.getOrCreateGauge(metricName)
		gauge.WithLabelValues(upsName).Set(f)
	}
}

func (w *MetricWriter) getOrCreateGauge(name string) *prometheus.GaugeVec {
	w.mu.Lock()
	defer w.mu.Unlock()

	if g, ok := w.gauges[name]; ok {
		return g
	}

	g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "NUT variable: " + name,
	}, []string{"ups"})
	w.reg.MustRegister(g)
	w.gauges[name] = g
	return g
}

func varToMetricName(nutVar string) string {
	s := strings.ReplaceAll(nutVar, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return "mjolnir_" + s
}
