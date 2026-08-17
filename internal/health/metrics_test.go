package health

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newTestWriter(t *testing.T) (*MetricWriter, *prometheus.Registry) {
	t.Helper()
	reg := prometheus.NewRegistry()
	return NewMetricWriter(reg), reg
}

func TestVarToMetricName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"battery.charge", "mjolnir_battery_charge"},
		{"battery.voltage.nominal", "mjolnir_battery_voltage_nominal"},
		{"input.voltage", "mjolnir_input_voltage"},
		{"ups.load", "mjolnir_ups_load"},
		{"output.frequency", "mjolnir_output_frequency"},
		{"battery.charge-low", "mjolnir_battery_charge_low"},
	}
	for _, tt := range tests {
		got := varToMetricName(tt.input)
		if got != tt.want {
			t.Errorf("varToMetricName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStatusFlags(t *testing.T) {
	w, reg := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.status": "OL CHRG",
	})

	val := testutil.ToFloat64(w.statusGauge.WithLabelValues("myups", "OL"))
	if val != 1 {
		t.Errorf("OL flag: got %v, want 1", val)
	}
	val = testutil.ToFloat64(w.statusGauge.WithLabelValues("myups", "CHRG"))
	if val != 1 {
		t.Errorf("CHRG flag: got %v, want 1", val)
	}
	val = testutil.ToFloat64(w.statusGauge.WithLabelValues("myups", "OB"))
	if val != 0 {
		t.Errorf("OB flag: got %v, want 0", val)
	}
	val = testutil.ToFloat64(w.statusGauge.WithLabelValues("myups", "LB"))
	if val != 0 {
		t.Errorf("LB flag: got %v, want 0", val)
	}

	count, err := testutil.GatherAndCount(reg, "mjolnir_ups_status")
	if err != nil {
		t.Fatal(err)
	}
	if count != len(statusFlags) {
		t.Errorf("expected %d status series, got %d", len(statusFlags), count)
	}
}

func TestStatusFlagsOnBattery(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.status": "OB LB DISCHRG",
	})

	for _, tc := range []struct {
		flag string
		want float64
	}{
		{"OB", 1}, {"LB", 1}, {"DISCHRG", 1},
		{"OL", 0}, {"CHRG", 0}, {"HB", 0},
	} {
		val := testutil.ToFloat64(w.statusGauge.WithLabelValues("myups", tc.flag))
		if val != tc.want {
			t.Errorf("flag %s: got %v, want %v", tc.flag, val, tc.want)
		}
	}
}

func TestStatusFlagsMissingStatus(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"battery.charge": "100",
	})

	for _, flag := range statusFlags {
		val := testutil.ToFloat64(w.statusGauge.WithLabelValues("myups", flag))
		if val != 0 {
			t.Errorf("flag %s should be 0 when ups.status missing, got %v", flag, val)
		}
	}
}

func TestDeviceInfo(t *testing.T) {
	w, reg := newTestWriter(t)

	w.Write("myups", map[string]string{
		"device.model":  "Smart-UPS 1500",
		"device.mfr":    "APC",
		"device.serial": "AS123",
		"device.type":   "ups",
	})

	val := testutil.ToFloat64(w.deviceInfo.WithLabelValues("myups", "Smart-UPS 1500", "APC", "AS123", "ups"))
	if val != 1 {
		t.Errorf("device_info: got %v, want 1", val)
	}

	count, err := testutil.GatherAndCount(reg, "mjolnir_device_info")
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 device_info series, got %d", count)
	}
}

func TestDeviceInfoPartial(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"device.model": "EcoFlow",
	})

	val := testutil.ToFloat64(w.deviceInfo.WithLabelValues("myups", "EcoFlow", "", "", ""))
	if val != 1 {
		t.Errorf("device_info partial: got %v, want 1", val)
	}
}

func TestDynamicGauges(t *testing.T) {
	w, reg := newTestWriter(t)

	w.Write("myups", map[string]string{
		"battery.charge":          "95.5",
		"input.voltage":           "230.1",
		"battery.voltage.nominal": "24.0",
		"ups.load":                "42",
		"ups.status":              "OL",
		"device.model":            "TestUPS",
		"driver.version":          "2.8.0",
		"ups.firmware":            "FW:2.0",
	})

	expected := map[string]float64{
		"mjolnir_battery_charge":          95.5,
		"mjolnir_input_voltage":           230.1,
		"mjolnir_battery_voltage_nominal": 24.0,
		"mjolnir_ups_load":               42,
	}

	for name, want := range expected {
		val := testutil.ToFloat64(w.gauges[name].WithLabelValues("myups"))
		if val != want {
			t.Errorf("%s: got %v, want %v", name, val, want)
		}
	}

	shouldNotExist := []string{
		"mjolnir_ups_status",
		"mjolnir_device_model",
		"mjolnir_driver_version",
		"mjolnir_ups_firmware",
	}
	for _, name := range shouldNotExist {
		if _, ok := w.gauges[name]; ok {
			t.Errorf("gauge %s should not exist in dynamic registry", name)
		}
	}

	gathered, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range gathered {
		if strings.HasPrefix(*mf.Name, "mjolnir_driver_") {
			t.Errorf("driver.* variable should be excluded, found %s", *mf.Name)
		}
	}
}

func TestDynamicGaugesNewVarOnLaterPoll(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"battery.charge": "100",
	})
	if _, ok := w.gauges["mjolnir_battery_charge"]; !ok {
		t.Fatal("expected mjolnir_battery_charge after first write")
	}
	if _, ok := w.gauges["mjolnir_battery_runtime"]; ok {
		t.Fatal("mjolnir_battery_runtime should not exist yet")
	}

	w.Write("myups", map[string]string{
		"battery.charge":  "99",
		"battery.runtime": "3600",
	})
	if _, ok := w.gauges["mjolnir_battery_runtime"]; !ok {
		t.Fatal("expected mjolnir_battery_runtime after second write")
	}
	val := testutil.ToFloat64(w.gauges["mjolnir_battery_runtime"].WithLabelValues("myups"))
	if val != 3600 {
		t.Errorf("battery_runtime: got %v, want 3600", val)
	}
}

func TestScrapeError(t *testing.T) {
	w, _ := newTestWriter(t)

	w.SetError("myups")
	val := testutil.ToFloat64(w.scrapeError.WithLabelValues("myups"))
	if val != 1 {
		t.Errorf("scrape_error after SetError: got %v, want 1", val)
	}

	w.Write("myups", map[string]string{"battery.charge": "100"})
	val = testutil.ToFloat64(w.scrapeError.WithLabelValues("myups"))
	if val != 0 {
		t.Errorf("scrape_error after Write: got %v, want 0", val)
	}
}

func TestWriteMultiUPS(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("ups1", map[string]string{
		"battery.charge": "100",
		"ups.status":     "OL",
	})
	w.Write("ups2", map[string]string{
		"battery.charge": "50",
		"ups.status":     "OB DISCHRG",
	})

	g := w.gauges["mjolnir_battery_charge"]
	v1 := testutil.ToFloat64(g.WithLabelValues("ups1"))
	v2 := testutil.ToFloat64(g.WithLabelValues("ups2"))
	if v1 != 100 {
		t.Errorf("ups1 battery_charge: got %v, want 100", v1)
	}
	if v2 != 50 {
		t.Errorf("ups2 battery_charge: got %v, want 50", v2)
	}

	ol1 := testutil.ToFloat64(w.statusGauge.WithLabelValues("ups1", "OL"))
	ol2 := testutil.ToFloat64(w.statusGauge.WithLabelValues("ups2", "OL"))
	ob2 := testutil.ToFloat64(w.statusGauge.WithLabelValues("ups2", "OB"))
	if ol1 != 1 {
		t.Errorf("ups1 OL: got %v, want 1", ol1)
	}
	if ol2 != 0 {
		t.Errorf("ups2 OL: got %v, want 0", ol2)
	}
	if ob2 != 1 {
		t.Errorf("ups2 OB: got %v, want 1", ob2)
	}
}

func TestWriteRealisticNUTOutput(t *testing.T) {
	w, reg := newTestWriter(t)

	vars := map[string]string{
		"battery.charge":          "100",
		"battery.charge.low":      "10",
		"battery.charge.warning":  "20",
		"battery.mfr.date":        "CPS",
		"battery.runtime":         "3660",
		"battery.runtime.low":     "300",
		"battery.type":            "PbAcid",
		"battery.voltage":         "13.5",
		"battery.voltage.nominal": "12.0",
		"device.mfr":              "CPS",
		"device.model":            "UPS CP1500",
		"device.serial":           "000000",
		"device.type":             "ups",
		"driver.name":             "usbhid-ups",
		"driver.parameter.pollfreq": "30",
		"driver.parameter.pollinterval": "2",
		"driver.parameter.port":  "auto",
		"driver.version":         "2.8.0",
		"driver.version.data":    "CyberPower HID 0.6",
		"driver.version.internal": "0.47",
		"input.transfer.high":    "260",
		"input.transfer.low":     "170",
		"input.voltage":          "230.0",
		"input.voltage.nominal":  "230",
		"output.voltage":         "230.0",
		"ups.beeper.status":      "enabled",
		"ups.delay.shutdown":     "20",
		"ups.delay.start":        "30",
		"ups.load":               "15",
		"ups.mfr":                "CPS",
		"ups.model":              "UPS CP1500",
		"ups.productid":          "0501",
		"ups.realpower.nominal":  "900",
		"ups.serial":             "000000",
		"ups.status":             "OL",
		"ups.test.result":        "No test initiated",
		"ups.timer.shutdown":     "-60",
		"ups.timer.start":        "-60",
		"ups.vendorid":           "0764",
	}

	w.Write("testups", vars)

	expectedGauges := []string{
		"mjolnir_battery_charge",
		"mjolnir_battery_charge_low",
		"mjolnir_battery_charge_warning",
		"mjolnir_battery_runtime",
		"mjolnir_battery_runtime_low",
		"mjolnir_battery_voltage",
		"mjolnir_battery_voltage_nominal",
		"mjolnir_input_transfer_high",
		"mjolnir_input_transfer_low",
		"mjolnir_input_voltage",
		"mjolnir_input_voltage_nominal",
		"mjolnir_output_voltage",
		"mjolnir_ups_delay_shutdown",
		"mjolnir_ups_delay_start",
		"mjolnir_ups_load",
		"mjolnir_ups_realpower_nominal",
		"mjolnir_ups_timer_shutdown",
		"mjolnir_ups_timer_start",
	}

	for _, name := range expectedGauges {
		if _, ok := w.gauges[name]; !ok {
			t.Errorf("expected gauge %s to exist", name)
		}
	}

	excludedPrefixes := []string{"mjolnir_driver_", "mjolnir_device_"}
	for name := range w.gauges {
		for _, prefix := range excludedPrefixes {
			if strings.HasPrefix(name, prefix) {
				t.Errorf("gauge %s should be excluded", name)
			}
		}
	}

	gathered, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	hasDeviceInfo := false
	hasStatus := false
	for _, mf := range gathered {
		if *mf.Name == "mjolnir_device_info" {
			hasDeviceInfo = true
		}
		if *mf.Name == "mjolnir_ups_status" {
			hasStatus = true
		}
	}
	if !hasDeviceInfo {
		t.Error("expected mjolnir_device_info metric")
	}
	if !hasStatus {
		t.Error("expected mjolnir_ups_status metric")
	}
}
