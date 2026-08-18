package health

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestEnumGaugeKnownValue(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.test.result": "OK",
	})

	val := testutil.ToFloat64(w.enumGauges["ups.test.result"].WithLabelValues("myups", "OK"))
	if val != 1 {
		t.Errorf("OK: got %v, want 1", val)
	}
	val = testutil.ToFloat64(w.enumGauges["ups.test.result"].WithLabelValues("myups", "Failed"))
	if val != 0 {
		t.Errorf("Failed: got %v, want 0", val)
	}
}

func TestEnumGaugeUnknownValue(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.test.result": "SomeNewValue",
	})

	for _, v := range enumRegistry["ups.test.result"].values {
		val := testutil.ToFloat64(w.enumGauges["ups.test.result"].WithLabelValues("myups", v))
		if val != 0 {
			t.Errorf("%s: got %v, want 0 for unknown enum value", v, val)
		}
	}
}

func TestEnumGaugeVariableAbsent(t *testing.T) {
	w, reg := newTestWriter(t)

	w.Write("myups", map[string]string{
		"battery.charge": "100",
	})

	gathered, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range gathered {
		if *mf.Name == "mjolnir_ups_test_result" {
			for _, m := range mf.Metric {
				for _, lp := range m.Label {
					if *lp.Name == "ups" && *lp.Value == "myups" {
						t.Error("ups.test.result metric should not be emitted when variable is absent")
					}
				}
			}
		}
	}
}

func TestEnumGaugeBatteryChargerStatus(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"battery.charger.status": "charging",
	})

	val := testutil.ToFloat64(w.enumGauges["battery.charger.status"].WithLabelValues("myups", "charging"))
	if val != 1 {
		t.Errorf("charging: got %v, want 1", val)
	}
	val = testutil.ToFloat64(w.enumGauges["battery.charger.status"].WithLabelValues("myups", "discharging"))
	if val != 0 {
		t.Errorf("discharging: got %v, want 0", val)
	}
}

func TestEnumGaugeInputVoltageStatus(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"input.voltage.status": "warning-low",
	})

	val := testutil.ToFloat64(w.enumGauges["input.voltage.status"].WithLabelValues("myups", "warning-low"))
	if val != 1 {
		t.Errorf("warning-low: got %v, want 1", val)
	}
	val = testutil.ToFloat64(w.enumGauges["input.voltage.status"].WithLabelValues("myups", "good"))
	if val != 0 {
		t.Errorf("good: got %v, want 0", val)
	}
}

func TestParseNUTDateFormats(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"2026-08-15 14:30:00", 1786804200},
		{"2024-03-01", 1709251200},
		{"2024/03/01", 1709251200},
		{"03/01/2024", 1709251200},
		{"1723729800", 1723729800},
	}
	for _, tt := range tests {
		got, err := parseNUTDate(tt.input)
		if err != nil {
			t.Errorf("parseNUTDate(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseNUTDate(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseNUTDateUnparseable(t *testing.T) {
	_, err := parseNUTDate("not set")
	if err == nil {
		t.Error("expected error for unparseable date")
	}
}

func TestTimestampGauge(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.test.date": "2024-03-01",
	})

	val := testutil.ToFloat64(w.timestampGauges["ups.test.date"].WithLabelValues("myups"))
	if val != 1709251200 {
		t.Errorf("ups.test.date: got %v, want 1709251200", val)
	}
}

func TestTimestampGaugeUnparseable(t *testing.T) {
	w, reg := newTestWriter(t)

	w.Write("myups", map[string]string{
		"battery.date": "not set",
	})

	gathered, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range gathered {
		if *mf.Name == "mjolnir_battery_date_seconds" {
			for _, m := range mf.Metric {
				for _, lp := range m.Label {
					if *lp.Name == "ups" && *lp.Value == "myups" {
						t.Error("timestamp metric should not be emitted for unparseable date")
					}
				}
			}
		}
	}
}

func TestInfoGaugeInitialValue(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.firmware":     "925.T2 .I",
		"ups.firmware.aux": "08.3",
	})

	val := testutil.ToFloat64(w.infoGauges["mjolnir_ups_firmware_info"].WithLabelValues("myups", "925.T2 .I", "08.3"))
	if val != 1 {
		t.Errorf("firmware info: got %v, want 1", val)
	}
}

func TestInfoGaugeValueChange(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"battery.type": "PbAcid",
	})

	val := testutil.ToFloat64(w.infoGauges["mjolnir_battery_type_info"].WithLabelValues("myups", "PbAcid"))
	if val != 1 {
		t.Errorf("PbAcid: got %v, want 1", val)
	}

	w.Write("myups", map[string]string{
		"battery.type": "LiIon",
	})

	val = testutil.ToFloat64(w.infoGauges["mjolnir_battery_type_info"].WithLabelValues("myups", "PbAcid"))
	if val != 0 {
		t.Errorf("PbAcid after change: got %v, want 0", val)
	}
	val = testutil.ToFloat64(w.infoGauges["mjolnir_battery_type_info"].WithLabelValues("myups", "LiIon"))
	if val != 1 {
		t.Errorf("LiIon: got %v, want 1", val)
	}
}

func TestInfoGaugeIdempotent(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{"ups.type": "online"})
	w.Write("myups", map[string]string{"ups.type": "online"})

	val := testutil.ToFloat64(w.infoGauges["mjolnir_ups_type_info"].WithLabelValues("myups", "online"))
	if val != 1 {
		t.Errorf("online: got %v, want 1", val)
	}
}

func TestAlarmActive(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.alarm": "Replace battery!",
	})

	val := testutil.ToFloat64(w.alarmActive.WithLabelValues("myups"))
	if val != 1 {
		t.Errorf("alarm_active: got %v, want 1", val)
	}
	val = testutil.ToFloat64(w.alarmInfo.WithLabelValues("myups", "Replace battery!"))
	if val != 1 {
		t.Errorf("alarm_info: got %v, want 1", val)
	}
}

func TestAlarmClears(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.alarm": "Replace battery!",
	})
	w.Write("myups", map[string]string{})

	val := testutil.ToFloat64(w.alarmActive.WithLabelValues("myups"))
	if val != 0 {
		t.Errorf("alarm_active after clear: got %v, want 0", val)
	}
	val = testutil.ToFloat64(w.alarmInfo.WithLabelValues("myups", "Replace battery!"))
	if val != 0 {
		t.Errorf("alarm_info after clear: got %v, want 0", val)
	}
}

func TestAlarmChangesText(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.alarm": "Replace battery!",
	})
	w.Write("myups", map[string]string{
		"ups.alarm": "Temperature high",
	})

	val := testutil.ToFloat64(w.alarmInfo.WithLabelValues("myups", "Replace battery!"))
	if val != 0 {
		t.Errorf("old alarm: got %v, want 0", val)
	}
	val = testutil.ToFloat64(w.alarmInfo.WithLabelValues("myups", "Temperature high"))
	if val != 1 {
		t.Errorf("new alarm: got %v, want 1", val)
	}
	val = testutil.ToFloat64(w.alarmActive.WithLabelValues("myups"))
	if val != 1 {
		t.Errorf("alarm_active: got %v, want 1", val)
	}
}

func TestIntegrationMixedVars(t *testing.T) {
	w, reg := newTestWriter(t)

	w.Write("myups", map[string]string{
		"battery.charge":         "95",
		"ups.status":             "OL",
		"ups.test.result":        "OK",
		"battery.charger.status": "floating",
		"ups.test.date":          "2024-03-01",
		"ups.firmware":           "FW:2.0",
		"ups.alarm":              "Replace battery!",
		"device.model":           "Smart-UPS",
		"driver.name":            "usbhid-ups",
		"input.voltage":          "230.0",
	})

	if _, ok := w.gauges["mjolnir_battery_charge"]; !ok {
		t.Error("expected numeric gauge mjolnir_battery_charge")
	}
	if _, ok := w.gauges["mjolnir_input_voltage"]; !ok {
		t.Error("expected numeric gauge mjolnir_input_voltage")
	}

	val := testutil.ToFloat64(w.enumGauges["ups.test.result"].WithLabelValues("myups", "OK"))
	if val != 1 {
		t.Errorf("enum ups.test.result OK: got %v, want 1", val)
	}

	val = testutil.ToFloat64(w.enumGauges["battery.charger.status"].WithLabelValues("myups", "floating"))
	if val != 1 {
		t.Errorf("enum battery.charger.status floating: got %v, want 1", val)
	}

	val = testutil.ToFloat64(w.timestampGauges["ups.test.date"].WithLabelValues("myups"))
	if val != 1709251200 {
		t.Errorf("timestamp ups.test.date: got %v, want 1709251200", val)
	}

	val = testutil.ToFloat64(w.infoGauges["mjolnir_ups_firmware_info"].WithLabelValues("myups", "FW:2.0", ""))
	if val != 1 {
		t.Errorf("info firmware: got %v, want 1", val)
	}

	val = testutil.ToFloat64(w.alarmActive.WithLabelValues("myups"))
	if val != 1 {
		t.Errorf("alarm active: got %v, want 1", val)
	}

	if _, ok := w.gauges["mjolnir_ups_test_result"]; ok {
		t.Error("string var ups.test.result should not create a dynamic gauge")
	}
	if _, ok := w.gauges["mjolnir_ups_firmware"]; ok {
		t.Error("string var ups.firmware should not create a dynamic gauge")
	}

	gathered, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range gathered {
		if *mf.Name == "mjolnir_driver_name" {
			t.Error("driver.* should be excluded")
		}
	}
}

func TestStringVarsNotInDynamicGauges(t *testing.T) {
	w, _ := newTestWriter(t)

	w.Write("myups", map[string]string{
		"ups.test.result":        "OK",
		"battery.charger.status": "charging",
		"ups.beeper.status":      "enabled",
		"input.sensitivity":      "high",
		"ups.test.date":          "2024-03-01",
		"battery.type":           "PbAcid",
		"ups.alarm":              "test",
	})

	excluded := []string{
		"mjolnir_ups_test_result",
		"mjolnir_battery_charger_status",
		"mjolnir_ups_beeper_status",
		"mjolnir_input_sensitivity",
		"mjolnir_ups_test_date",
		"mjolnir_battery_type",
		"mjolnir_ups_alarm",
	}
	for _, name := range excluded {
		if _, ok := w.gauges[name]; ok {
			t.Errorf("string var %s should not appear in dynamic gauges", name)
		}
	}
}
