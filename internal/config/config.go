package config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type UPSConfig struct {
	Name         string
	Driver       string
	Port         string
	Desc         string
	Serial       string
	VendorID     string
	PollInterval string
	SDOrder      string
	Extra        map[string]string
}

type Config struct {
	UPSUnits   []UPSConfig
	User       string
	Password   string
	SecretName string
	Server     string
	Listen     string
	MaxAge     string
}

var upsNameRe = regexp.MustCompile(`NUT_UPS_(\d+)_NAME`)

func Load() (*Config, error) {
	cfg := &Config{
		User:       envOr("NUT_USER", "admin"),
		SecretName: envOr("NUT_SECRET_NAME", "nut-password"),
		Server:     envOr("NUT_SERVER", "primary"),
		Listen:     envOr("NUT_LISTEN", "0.0.0.0"),
		MaxAge:     envOr("NUT_MAXAGE", "15"),
	}

	indices := discoverUPSIndices()
	if len(indices) == 0 {
		return nil, fmt.Errorf("no UPS units found (set NUT_UPS_<n>_NAME)")
	}

	for _, idx := range indices {
		prefix := fmt.Sprintf("NUT_UPS_%d_", idx)
		ups := UPSConfig{
			Name:         os.Getenv(prefix + "NAME"),
			Driver:       envOr(prefix+"DRIVER", "usbhid-ups"),
			Port:         envOr(prefix+"PORT", "auto"),
			Desc:         envOr(prefix+"DESC", "UPS"),
			Serial:       os.Getenv(prefix + "SERIAL"),
			VendorID:     os.Getenv(prefix + "VENDORID"),
			PollInterval: os.Getenv(prefix + "POLLINTERVAL"),
			SDOrder:      os.Getenv(prefix + "SDORDER"),
			Extra:        parseExtra(os.Getenv(prefix + "EXTRA")),
		}
		cfg.UPSUnits = append(cfg.UPSUnits, ups)
	}

	cfg.Password = resolvePassword(cfg.SecretName)

	log.Printf("[mjolnir] loaded %d UPS unit(s), user=%s, server=%s", len(cfg.UPSUnits), cfg.User, cfg.Server)
	return cfg, nil
}

func discoverUPSIndices() []int {
	var indices []int
	seen := make(map[int]bool)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		matches := upsNameRe.FindStringSubmatch(parts[0])
		if matches == nil {
			continue
		}
		idx, err := strconv.Atoi(matches[1])
		if err != nil || seen[idx] {
			continue
		}
		seen[idx] = true
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	return indices
}

func resolvePassword(secretName string) string {
	secretPath := fmt.Sprintf("/run/secrets/%s", secretName)
	data, err := os.ReadFile(secretPath)
	if err == nil {
		pw := strings.TrimSpace(string(data))
		if pw != "" {
			log.Printf("[mjolnir] password loaded from Docker secret %s", secretName)
			return pw
		}
	}

	if pw := os.Getenv("NUT_PASSWORD"); pw != "" {
		return pw
	}

	log.Printf("[mjolnir] WARNING: using default password, set NUT_PASSWORD or mount a Docker secret")
	return "changeme"
}

func parseExtra(s string) map[string]string {
	m := make(map[string]string)
	if s == "" {
		return m
	}
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		k, v, ok := strings.Cut(pair, "=")
		if !ok || k == "" {
			continue
		}
		m[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return m
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
