package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testConfig() *Config {
	return &Config{
		UPSUnits: []UPSConfig{
			{Name: "ups1", Driver: "usbhid-ups", Port: "auto", Desc: "Main UPS"},
			{Name: "ups2", Driver: "snmp-ups", Port: "192.168.1.100", Desc: "Remote UPS", Extra: map[string]string{}},
		},
		User:     "admin",
		Password: "testpass",
		Server:   "primary",
		Listen:   "0.0.0.0",
		MaxAge:   "15",
	}
}

func TestGenerateUpsConf(t *testing.T) {
	cfg := testConfig()
	runDir := t.TempDir()

	if err := generateUpsConf(cfg, runDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(runDir, "ups.conf"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	expects := []string{
		"maxretry = 3",
		"[ups1]",
		"  driver = usbhid-ups",
		"  port = auto",
		`  desc = "Main UPS"`,
		"[ups2]",
		"  driver = snmp-ups",
		"  port = 192.168.1.100",
		`  desc = "Remote UPS"`,
	}
	for _, exp := range expects {
		if !strings.Contains(content, exp) {
			t.Errorf("ups.conf missing %q", exp)
		}
	}
}

func TestGenerateUpsdConf(t *testing.T) {
	cfg := testConfig()
	runDir := t.TempDir()
	localDir := t.TempDir()

	if err := generateUpsdConf(cfg, runDir, localDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(runDir, "upsd.conf"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "LISTEN 0.0.0.0 3493") {
		t.Error("upsd.conf missing LISTEN line")
	}
	if !strings.Contains(content, "MAXAGE 15") {
		t.Error("upsd.conf missing MAXAGE line")
	}
}

func TestGenerateUpsdUsers(t *testing.T) {
	cfg := testConfig()
	runDir := t.TempDir()

	if err := generateUpsdUsers(cfg, runDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(runDir, "upsd.users"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	expects := []string{
		"[admin]",
		"  password = testpass",
		"  upsmon primary",
	}
	for _, exp := range expects {
		if !strings.Contains(content, exp) {
			t.Errorf("upsd.users missing %q", exp)
		}
	}
}

func TestGenerateUpsmonConf(t *testing.T) {
	cfg := testConfig()
	runDir := t.TempDir()

	if err := generateUpsmonConf(cfg, runDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(runDir, "upsmon.conf"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	expects := []string{
		"MONITOR ups1@localhost 1 admin testpass primary",
		"MONITOR ups2@localhost 1 admin testpass primary",
		"RUN_AS_USER nut",
	}
	for _, exp := range expects {
		if !strings.Contains(content, exp) {
			t.Errorf("upsmon.conf missing %q", exp)
		}
	}
}

func TestMountedConfigPassthrough(t *testing.T) {
	cfg := testConfig()
	runDir := t.TempDir()
	localDir := t.TempDir()

	mounted := filepath.Join(localDir, "upsd.conf")
	original := "LISTEN 127.0.0.1 3493\nMAXAGE 20\n"
	if err := os.WriteFile(mounted, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateUpsdConf(cfg, runDir, localDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(runDir, "upsd.conf"))
	if err != nil {
		t.Fatal(err)
	}

	// MaxAge is default (15), so mounted content passes through unchanged
	if string(data) != original {
		t.Errorf("mounted config not passed through:\ngot:  %q\nwant: %q", string(data), original)
	}
}

func TestUpsdConfMaxageOverride(t *testing.T) {
	cfg := testConfig()
	cfg.MaxAge = "30"
	runDir := t.TempDir()
	localDir := t.TempDir()

	mounted := filepath.Join(localDir, "upsd.conf")
	if err := os.WriteFile(mounted, []byte("LISTEN 127.0.0.1 3493\nMAXAGE 20\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateUpsdConf(cfg, runDir, localDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(runDir, "upsd.conf"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "MAXAGE 30") {
		t.Errorf("MAXAGE not overridden, got:\n%s", content)
	}
	if strings.Contains(content, "MAXAGE 20") {
		t.Error("old MAXAGE 20 still present")
	}
}
