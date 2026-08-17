package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Generate(cfg *Config, runDir, localDir string) error {
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return fmt.Errorf("create run dir: %w", err)
	}

	generators := []struct {
		name string
		fn   func() error
	}{
		{"nut.conf", func() error { return generateNutConf(runDir) }},
		{"ups.conf", func() error { return generateOrCopy("ups.conf", localDir, runDir, func() error { return generateUpsConf(cfg, runDir) }) }},
		{"upsd.conf", func() error { return generateUpsdConf(cfg, runDir, localDir) }},
		{"upsd.users", func() error { return generateOrCopy("upsd.users", localDir, runDir, func() error { return generateUpsdUsers(cfg, runDir) }) }},
		{"upsmon.conf", func() error { return generateOrCopy("upsmon.conf", localDir, runDir, func() error { return generateUpsmonConf(cfg, runDir) }) }},
	}

	for _, g := range generators {
		if err := g.fn(); err != nil {
			return fmt.Errorf("generate %s: %w", g.name, err)
		}
	}
	return nil
}

func generateNutConf(runDir string) error {
	return os.WriteFile(filepath.Join(runDir, "nut.conf"), []byte("MODE=netserver\n"), 0640)
}

func generateUpsConf(cfg *Config, runDir string) error {
	var b strings.Builder
	b.WriteString("maxretry = 3\n")

	for _, ups := range cfg.UPSUnits {
		fmt.Fprintf(&b, "\n[%s]\n", ups.Name)
		fmt.Fprintf(&b, "  driver = %s\n", ups.Driver)
		fmt.Fprintf(&b, "  port = %s\n", ups.Port)
		fmt.Fprintf(&b, "  desc = \"%s\"\n", ups.Desc)

		if ups.Serial != "" {
			fmt.Fprintf(&b, "  serial = \"%s\"\n", ups.Serial)
		}
		if ups.VendorID != "" {
			fmt.Fprintf(&b, "  vendorid = %s\n", ups.VendorID)
		}
		if ups.PollInterval != "" {
			fmt.Fprintf(&b, "  pollinterval = %s\n", ups.PollInterval)
		}
		if ups.SDOrder != "" {
			fmt.Fprintf(&b, "  sdorder = %s\n", ups.SDOrder)
		}

		keys := make([]string, 0, len(ups.Extra))
		for k := range ups.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "  %s = %s\n", k, ups.Extra[k])
		}
	}

	return os.WriteFile(filepath.Join(runDir, "ups.conf"), []byte(b.String()), 0640)
}

func generateUpsdConf(cfg *Config, runDir, localDir string) error {
	mounted := filepath.Join(localDir, "upsd.conf")
	dst := filepath.Join(runDir, "upsd.conf")

	if _, err := os.Stat(mounted); err == nil {
		data, err := os.ReadFile(mounted)
		if err != nil {
			return err
		}
		content := string(data)

		if cfg.MaxAge != "15" {
			content = applyMaxageOverride(content, cfg.MaxAge)
		}

		if !strings.Contains(strings.ToUpper(content), "LISTEN") {
			content = strings.TrimRight(content, "\n") + fmt.Sprintf("\nLISTEN %s 3493\n", cfg.Listen)
		}

		return os.WriteFile(dst, []byte(content), 0640)
	}

	content := fmt.Sprintf("LISTEN %s 3493\nMAXAGE %s\n", cfg.Listen, cfg.MaxAge)
	return os.WriteFile(dst, []byte(content), 0640)
}

func applyMaxageOverride(content, maxage string) string {
	lines := strings.Split(content, "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(trimmed), "MAXAGE") {
			lines[i] = fmt.Sprintf("MAXAGE %s", maxage)
			found = true
		}
	}
	if !found {
		lines = append(lines, fmt.Sprintf("MAXAGE %s", maxage))
	}
	return strings.Join(lines, "\n")
}

func generateUpsdUsers(cfg *Config, runDir string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s]\n", cfg.User)
	fmt.Fprintf(&b, "  password = %s\n", cfg.Password)
	fmt.Fprintf(&b, "  upsmon %s\n", cfg.Server)

	return os.WriteFile(filepath.Join(runDir, "upsd.users"), []byte(b.String()), 0640)
}

func generateUpsmonConf(cfg *Config, runDir string) error {
	var b strings.Builder
	for _, ups := range cfg.UPSUnits {
		fmt.Fprintf(&b, "MONITOR %s@localhost 1 %s %s %s\n", ups.Name, cfg.User, cfg.Password, cfg.Server)
	}
	b.WriteString("RUN_AS_USER nut\n")

	return os.WriteFile(filepath.Join(runDir, "upsmon.conf"), []byte(b.String()), 0640)
}

func generateOrCopy(name, localDir, runDir string, generate func() error) error {
	mounted := filepath.Join(localDir, name)
	if _, err := os.Stat(mounted); err == nil {
		log.Printf("[mjolnir] using mounted %s", name)
		return copyFile(mounted, filepath.Join(runDir, name))
	}
	log.Printf("[mjolnir] generating %s from env", name)
	return generate()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Chmod(0640)
}
