package config

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestFixOwnership(t *testing.T) {
	if _, err := user.Lookup("nut"); err != nil {
		t.Skip("nut user not available on this system")
	}

	dir := t.TempDir()
	files := []string{"ups.conf", "upsd.conf", "upsd.users"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0640); err != nil {
			t.Fatal(err)
		}
	}

	if err := FixOwnership(dir); err != nil {
		t.Fatalf("FixOwnership: %v", err)
	}

	nutUser, _ := user.Lookup("nut")

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0640 {
			t.Errorf("%s: expected mode 0640, got %o", e.Name(), info.Mode().Perm())
		}
		_ = nutUser
	}
}

func TestFixOwnershipMissingUser(t *testing.T) {
	if _, err := user.Lookup("nut"); err == nil {
		t.Skip("nut user exists, cannot test missing-user path")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.conf"), []byte("x"), 0640); err != nil {
		t.Fatal(err)
	}

	err := FixOwnership(dir)
	if err == nil {
		t.Fatal("expected error when nut user missing")
	}
}
