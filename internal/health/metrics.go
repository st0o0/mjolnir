package health

import (
	"strconv"
	"strings"
	"sync"
	"time"

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

type enumDef struct {
	metricName string
	label      string
	values     []string
}

var sharedStatusValues = []string{"critical-low", "warning-low", "good", "warning-high", "critical-high"}

var enumRegistry = map[string]enumDef{
	"ups.test.result":         {"mjolnir_ups_test_result", "result", []string{"OK", "Failed", "InProgress", "Aborted", "NoTestInit", "BATTdetach"}},
	"battery.charger.status":  {"mjolnir_battery_charger_status", "status", []string{"charging", "discharging", "floating", "resting"}},
	"ups.beeper.status":       {"mjolnir_ups_beeper_status", "status", []string{"enabled", "disabled", "muted"}},
	"input.sensitivity":       {"mjolnir_input_sensitivity", "sensitivity", []string{"low", "medium", "high", "auto"}},
	"input.transfer.reason":   {"mjolnir_input_transfer_reason", "reason", []string{"noTransfer", "highLineVoltage", "brownout", "selfTest", "forcedReboot", "inputFreqOutOfRange", "inputVoltageOutOfRange"}},
	"input.voltage.status":    {"mjolnir_input_voltage_status", "status", sharedStatusValues},
	"input.current.status":    {"mjolnir_input_current_status", "status", sharedStatusValues},
	"input.frequency.status":  {"mjolnir_input_frequency_status", "status", sharedStatusValues},
}

var timestampRegistry = map[string]string{
	"ups.test.date":            "mjolnir_ups_test_date_seconds",
	"battery.date":             "mjolnir_battery_date_seconds",
	"battery.date.maintenance": "mjolnir_battery_date_maintenance_seconds",
	"battery.mfr.date":         "mjolnir_battery_mfr_date_seconds",
	"ups.mfr.date":             "mjolnir_ups_mfr_date_seconds",
}

type infoDef struct {
	metricName string
	nutVars    []string
	labels     []string
}

var infoRegistry = []infoDef{
	{"mjolnir_ups_firmware_info", []string{"ups.firmware", "ups.firmware.aux"}, []string{"version", "aux"}},
	{"mjolnir_battery_type_info", []string{"battery.type"}, []string{"type"}},
	{"mjolnir_ups_type_info", []string{"ups.type"}, []string{"type"}},
}

var stringRegistrySet map[string]bool

func init() {
	stringRegistrySet = make(map[string]bool)
	for k := range enumRegistry {
		stringRegistrySet[k] = true
	}
	for k := range timestampRegistry {
		stringRegistrySet[k] = true
	}
	for _, info := range infoRegistry {
		for _, v := range info.nutVars {
			stringRegistrySet[v] = true
		}
	}
	stringRegistrySet["ups.alarm"] = true
}

type MetricWriter struct {
	mu     sync.Mutex
	gauges map[string]*prometheus.GaugeVec
	reg    prometheus.Registerer

	statusGauge *prometheus.GaugeVec
	deviceInfo  *prometheus.GaugeVec
	scrapeError *prometheus.GaugeVec

	enumGauges      map[string]*prometheus.GaugeVec
	timestampGauges map[string]*prometheus.GaugeVec
	infoGauges      map[string]*prometheus.GaugeVec
	alarmActive     *prometheus.GaugeVec
	alarmInfo       *prometheus.GaugeVec

	prevInfoValues map[string]string
	prevAlarm      map[string]string
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

	enumGauges := make(map[string]*prometheus.GaugeVec)
	for nutVar, def := range enumRegistry {
		g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: def.metricName,
			Help: "NUT enum variable: " + nutVar,
		}, []string{"ups", def.label})
		enumGauges[nutVar] = g
	}

	timestampGauges := make(map[string]*prometheus.GaugeVec)
	for nutVar, metricName := range timestampRegistry {
		g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: metricName,
			Help: "NUT timestamp variable: " + nutVar,
		}, []string{"ups"})
		timestampGauges[nutVar] = g
	}

	infoGauges := make(map[string]*prometheus.GaugeVec)
	for _, def := range infoRegistry {
		g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: def.metricName,
			Help: "NUT info variable",
		}, append([]string{"ups"}, def.labels...))
		infoGauges[def.metricName] = g
	}

	alarmActive := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_alarm_active",
		Help: "1 if UPS has an active alarm, 0 otherwise",
	}, []string{"ups"})

	alarmInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mjolnir_ups_alarm_info",
		Help: "UPS alarm text as label",
	}, []string{"ups", "alarm"})

	toRegister := []prometheus.Collector{statusGauge, deviceInfo, scrapeError, alarmActive, alarmInfo}
	for _, g := range enumGauges {
		toRegister = append(toRegister, g)
	}
	for _, g := range timestampGauges {
		toRegister = append(toRegister, g)
	}
	for _, g := range infoGauges {
		toRegister = append(toRegister, g)
	}
	reg.MustRegister(toRegister...)

	return &MetricWriter{
		gauges:          make(map[string]*prometheus.GaugeVec),
		reg:             reg,
		statusGauge:     statusGauge,
		deviceInfo:      deviceInfo,
		scrapeError:     scrapeError,
		enumGauges:      enumGauges,
		timestampGauges: timestampGauges,
		infoGauges:      infoGauges,
		alarmActive:     alarmActive,
		alarmInfo:       alarmInfo,
		prevInfoValues:  make(map[string]string),
		prevAlarm:       make(map[string]string),
	}
}

func (w *MetricWriter) Write(upsName string, vars map[string]string) {
	w.scrapeError.WithLabelValues(upsName).Set(0)
	w.writeStatus(upsName, vars)
	w.writeDeviceInfo(upsName, vars)
	w.writeEnumGauges(upsName, vars)
	w.writeTimestampGauges(upsName, vars)
	w.writeInfoGauges(upsName, vars)
	w.writeAlarm(upsName, vars)
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

func (w *MetricWriter) writeEnumGauges(upsName string, vars map[string]string) {
	for nutVar, def := range enumRegistry {
		val, ok := vars[nutVar]
		if !ok {
			continue
		}
		for _, known := range def.values {
			v := 0.0
			if known == val {
				v = 1.0
			}
			w.enumGauges[nutVar].WithLabelValues(upsName, known).Set(v)
		}
	}
}

func (w *MetricWriter) writeTimestampGauges(upsName string, vars map[string]string) {
	for nutVar, gauge := range w.timestampGauges {
		val, ok := vars[nutVar]
		if !ok {
			continue
		}
		ts, err := parseNUTDate(val)
		if err != nil {
			continue
		}
		gauge.WithLabelValues(upsName).Set(float64(ts))
	}
}

func (w *MetricWriter) writeInfoGauges(upsName string, vars map[string]string) {
	for _, def := range infoRegistry {
		hasAny := false
		labelVals := make([]string, len(def.nutVars))
		for i, nutVar := range def.nutVars {
			if v, ok := vars[nutVar]; ok {
				labelVals[i] = v
				hasAny = true
			}
		}
		if !hasAny {
			continue
		}

		key := upsName + ":" + def.metricName
		newVal := strings.Join(labelVals, "\x00")
		if prev, ok := w.prevInfoValues[key]; ok && prev != newVal {
			prevLabels := strings.Split(prev, "\x00")
			w.infoGauges[def.metricName].WithLabelValues(append([]string{upsName}, prevLabels...)...).Set(0)
		}
		w.prevInfoValues[key] = newVal
		w.infoGauges[def.metricName].WithLabelValues(append([]string{upsName}, labelVals...)...).Set(1)
	}
}

func (w *MetricWriter) writeAlarm(upsName string, vars map[string]string) {
	alarm := vars["ups.alarm"]

	if alarm == "" {
		w.alarmActive.WithLabelValues(upsName).Set(0)
		if prev, ok := w.prevAlarm[upsName]; ok && prev != "" {
			w.alarmInfo.WithLabelValues(upsName, prev).Set(0)
			w.prevAlarm[upsName] = ""
		}
		return
	}

	w.alarmActive.WithLabelValues(upsName).Set(1)
	if prev, ok := w.prevAlarm[upsName]; ok && prev != alarm && prev != "" {
		w.alarmInfo.WithLabelValues(upsName, prev).Set(0)
	}
	w.prevAlarm[upsName] = alarm
	w.alarmInfo.WithLabelValues(upsName, alarm).Set(1)
}

func (w *MetricWriter) writeDynamicGauges(upsName string, vars map[string]string) {
	for nutVar, val := range vars {
		if nutVar == "ups.status" {
			continue
		}
		if strings.HasPrefix(nutVar, "device.") || strings.HasPrefix(nutVar, "driver.") {
			continue
		}
		if stringRegistrySet[nutVar] {
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

var dateFormats = []string{
	"2006-01-02 15:04:05",
	"2006-01-02",
	"2006/01/02",
	"01/02/2006",
}

func parseNUTDate(s string) (int64, error) {
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		return ts, nil
	}
	for _, fmt := range dateFormats {
		if t, err := time.Parse(fmt, s); err == nil {
			return t.Unix(), nil
		}
	}
	return 0, &strconv.NumError{Func: "parseNUTDate", Num: s, Err: strconv.ErrSyntax}
}
