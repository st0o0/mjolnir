package nut

import (
	"testing"
)

func TestParseUpscOutput(t *testing.T) {
	input := `battery.charge: 100
battery.voltage: 13.50
device.type: ups
input.voltage: 230.0
output.voltage: 230.0
ups.load: 15
ups.status: OL`

	result := parseUpscOutput(input)

	expected := map[string]string{
		"battery.charge":  "100",
		"battery.voltage": "13.50",
		"device.type":     "ups",
		"input.voltage":   "230.0",
		"output.voltage":  "230.0",
		"ups.load":        "15",
		"ups.status":      "OL",
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(result))
	}

	for k, want := range expected {
		got, ok := result[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("key %q: expected %q, got %q", k, want, got)
		}
	}
}

func TestParseUpscOutputEmpty(t *testing.T) {
	result := parseUpscOutput("")
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestParseUpscOutputMalformedLines(t *testing.T) {
	input := `battery.charge: 100
malformed-no-colon
ups.status: OL
another bad line`

	result := parseUpscOutput(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(result), result)
	}
	if result["battery.charge"] != "100" {
		t.Errorf("battery.charge: expected %q, got %q", "100", result["battery.charge"])
	}
	if result["ups.status"] != "OL" {
		t.Errorf("ups.status: expected %q, got %q", "OL", result["ups.status"])
	}
}
