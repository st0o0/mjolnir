package config

import (
	"testing"
)

func TestLoadSingleUPS(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.UPSUnits) != 1 {
		t.Fatalf("expected 1 UPS, got %d", len(cfg.UPSUnits))
	}
	ups := cfg.UPSUnits[0]
	if ups.Name != "ups" {
		t.Errorf("name = %q, want %q", ups.Name, "ups")
	}
	if ups.Driver != "usbhid-ups" {
		t.Errorf("driver = %q, want %q", ups.Driver, "usbhid-ups")
	}
	if ups.Port != "auto" {
		t.Errorf("port = %q, want %q", ups.Port, "auto")
	}
	if ups.Desc != "UPS" {
		t.Errorf("desc = %q, want %q", ups.Desc, "UPS")
	}
}

func TestLoadMultiUPS(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups1")
	t.Setenv("NUT_UPS_3_NAME", "ups3")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.UPSUnits) != 2 {
		t.Fatalf("expected 2 UPS units, got %d", len(cfg.UPSUnits))
	}
	if cfg.UPSUnits[0].Name != "ups1" {
		t.Errorf("first UPS = %q, want %q", cfg.UPSUnits[0].Name, "ups1")
	}
	if cfg.UPSUnits[1].Name != "ups3" {
		t.Errorf("second UPS = %q, want %q", cfg.UPSUnits[1].Name, "ups3")
	}
}

func TestLoadExtraExpansion(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_UPS_1_EXTRA", "offdelay=30,ondelay=60")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	extra := cfg.UPSUnits[0].Extra
	if extra["offdelay"] != "30" {
		t.Errorf("offdelay = %q, want %q", extra["offdelay"], "30")
	}
	if extra["ondelay"] != "60" {
		t.Errorf("ondelay = %q, want %q", extra["ondelay"], "60")
	}
}

func TestPasswordFromEnv(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_PASSWORD", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Password != "secret" {
		t.Errorf("password = %q, want %q", cfg.Password, "secret")
	}
}

func TestPasswordDefault(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Password != "changeme" {
		t.Errorf("password = %q, want %q", cfg.Password, "changeme")
	}
}

func TestParseExtra(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]string
	}{
		{"", map[string]string{}},
		{"key=val", map[string]string{"key": "val"}},
		{"a=1,b=2", map[string]string{"a": "1", "b": "2"}},
		{" a = 1 , b = 2 ", map[string]string{"a": "1", "b": "2"}},
		{"noequals", map[string]string{}},
		{"=empty", map[string]string{}},
	}

	for _, tt := range tests {
		got := parseExtra(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("parseExtra(%q) len = %d, want %d", tt.input, len(got), len(tt.want))
			continue
		}
		for k, v := range tt.want {
			if got[k] != v {
				t.Errorf("parseExtra(%q)[%q] = %q, want %q", tt.input, k, got[k], v)
			}
		}
	}
}
