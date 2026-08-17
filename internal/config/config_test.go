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

func TestLegacyPasswordFromEnv(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_PASSWORD", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(cfg.Users))
	}
	if cfg.Users[0].Password != "secret" {
		t.Errorf("password = %q, want %q", cfg.Users[0].Password, "secret")
	}
}

func TestLegacyPasswordDefault(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Users[0].Password != "changeme" {
		t.Errorf("password = %q, want %q", cfg.Users[0].Password, "changeme")
	}
}

func TestLegacyUserDefaults(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(cfg.Users))
	}
	u := cfg.Users[0]
	if u.Name != "admin" {
		t.Errorf("name = %q, want %q", u.Name, "admin")
	}
	if u.Upsmon != "primary" {
		t.Errorf("upsmon = %q, want %q", u.Upsmon, "primary")
	}
}

func TestMultiUserDiscovery(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_USER_1_NAME", "monitor")
	t.Setenv("NUT_USER_1_PASSWORD", "pass1")
	t.Setenv("NUT_USER_1_UPSMON", "primary")
	t.Setenv("NUT_USER_2_NAME", "admin")
	t.Setenv("NUT_USER_2_PASSWORD", "pass2")
	t.Setenv("NUT_USER_2_ACTIONS", "SET,FSD")
	t.Setenv("NUT_USER_2_INSTCMDS", "ALL")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(cfg.Users))
	}

	u1 := cfg.Users[0]
	if u1.Name != "monitor" {
		t.Errorf("user 1 name = %q, want %q", u1.Name, "monitor")
	}
	if u1.Upsmon != "primary" {
		t.Errorf("user 1 upsmon = %q, want %q", u1.Upsmon, "primary")
	}

	u2 := cfg.Users[1]
	if u2.Name != "admin" {
		t.Errorf("user 2 name = %q, want %q", u2.Name, "admin")
	}
	if len(u2.Actions) != 2 || u2.Actions[0] != "SET" || u2.Actions[1] != "FSD" {
		t.Errorf("user 2 actions = %v, want [SET FSD]", u2.Actions)
	}
	if len(u2.Instcmds) != 1 || u2.Instcmds[0] != "ALL" {
		t.Errorf("user 2 instcmds = %v, want [ALL]", u2.Instcmds)
	}
}

func TestSparseUserIndices(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_USER_1_NAME", "first")
	t.Setenv("NUT_USER_1_PASSWORD", "pass1")
	t.Setenv("NUT_USER_1_UPSMON", "primary")
	t.Setenv("NUT_USER_5_NAME", "fifth")
	t.Setenv("NUT_USER_5_PASSWORD", "pass5")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(cfg.Users))
	}
	if cfg.Users[0].Name != "first" {
		t.Errorf("first user = %q, want %q", cfg.Users[0].Name, "first")
	}
	if cfg.Users[1].Name != "fifth" {
		t.Errorf("second user = %q, want %q", cfg.Users[1].Name, "fifth")
	}
}

func TestIndexedUsersOverrideLegacy(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_USER", "legacy-admin")
	t.Setenv("NUT_PASSWORD", "legacy-pass")
	t.Setenv("NUT_USER_1_NAME", "indexed-monitor")
	t.Setenv("NUT_USER_1_PASSWORD", "indexed-pass")
	t.Setenv("NUT_USER_1_UPSMON", "primary")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(cfg.Users))
	}
	if cfg.Users[0].Name != "indexed-monitor" {
		t.Errorf("user name = %q, want %q", cfg.Users[0].Name, "indexed-monitor")
	}
}

func TestDuplicateUserNames(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_USER_1_NAME", "admin")
	t.Setenv("NUT_USER_1_PASSWORD", "pass1")
	t.Setenv("NUT_USER_2_NAME", "admin")
	t.Setenv("NUT_USER_2_PASSWORD", "pass2")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for duplicate user names")
	}
}

func TestInvalidAction(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_USER_1_NAME", "admin")
	t.Setenv("NUT_USER_1_PASSWORD", "pass1")
	t.Setenv("NUT_USER_1_ACTIONS", "INVALID")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestInvalidUpsmonRole(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_USER_1_NAME", "admin")
	t.Setenv("NUT_USER_1_PASSWORD", "pass1")
	t.Setenv("NUT_USER_1_UPSMON", "master")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid upsmon role")
	}
}

func TestMissingPasswordMultiUser(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_USER_1_NAME", "admin")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing password in multi-user mode")
	}
}

func TestActionsCaseInsensitive(t *testing.T) {
	t.Setenv("NUT_UPS_1_NAME", "ups")
	t.Setenv("NUT_USER_1_NAME", "admin")
	t.Setenv("NUT_USER_1_PASSWORD", "pass1")
	t.Setenv("NUT_USER_1_ACTIONS", "set,fsd")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Users[0].Actions[0] != "SET" || cfg.Users[0].Actions[1] != "FSD" {
		t.Errorf("actions = %v, want [SET FSD]", cfg.Users[0].Actions)
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
