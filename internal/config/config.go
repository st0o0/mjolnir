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

type UserConfig struct {
	Name       string
	Password   string
	SecretName string
	Upsmon     string
	Actions    []string
	Instcmds   []string
}

type Config struct {
	UPSUnits []UPSConfig
	Users    []UserConfig
	Listen   string
	MaxAge   string
}

var (
	upsNameRe  = regexp.MustCompile(`NUT_UPS_(\d+)_NAME`)
	userNameRe = regexp.MustCompile(`NUT_USER_(\d+)_NAME`)
)

func Load() (*Config, error) {
	cfg := &Config{
		Listen: envOr("NUT_LISTEN", "0.0.0.0"),
		MaxAge: envOr("NUT_MAXAGE", "15"),
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

	users, err := loadUsers()
	if err != nil {
		return nil, err
	}
	cfg.Users = users

	log.Printf("[mjolnir] loaded %d UPS unit(s), %d user(s)", len(cfg.UPSUnits), len(cfg.Users))
	return cfg, nil
}

func loadUsers() ([]UserConfig, error) {
	userIndices := discoverUserIndices()

	if len(userIndices) > 0 {
		if hasLegacyUserVars() {
			log.Printf("[mjolnir] WARNING: NUT_USER_<n>_* vars found, ignoring legacy NUT_USER/NUT_PASSWORD/NUT_SERVER")
		}
		return loadIndexedUsers(userIndices)
	}

	return loadLegacyUser()
}

func loadIndexedUsers(indices []int) ([]UserConfig, error) {
	seen := make(map[string]bool)
	var users []UserConfig

	for _, idx := range indices {
		prefix := fmt.Sprintf("NUT_USER_%d_", idx)
		name := os.Getenv(prefix + "NAME")

		if seen[name] {
			return nil, fmt.Errorf("duplicate user name %q (NUT_USER_%d_NAME)", name, idx)
		}
		seen[name] = true

		user := UserConfig{
			Name:       name,
			SecretName: os.Getenv(prefix + "SECRET_NAME"),
			Upsmon:     strings.ToLower(os.Getenv(prefix + "UPSMON")),
			Actions:    parseCSV(os.Getenv(prefix + "ACTIONS")),
			Instcmds:   parseCSV(os.Getenv(prefix + "INSTCMDS")),
		}

		pw, err := resolveUserPassword(idx, &user)
		if err != nil {
			return nil, err
		}
		user.Password = pw

		if err := validateUser(&user, idx); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func loadLegacyUser() ([]UserConfig, error) {
	user := UserConfig{
		Name:   envOr("NUT_USER", "admin"),
		Upsmon: envOr("NUT_SERVER", "primary"),
	}

	secretName := envOr("NUT_SECRET_NAME", "nut-password")
	user.Password = resolveLegacyPassword(secretName)

	return []UserConfig{user}, nil
}

func discoverUPSIndices() []int {
	return discoverIndices(upsNameRe)
}

func discoverUserIndices() []int {
	return discoverIndices(userNameRe)
}

func discoverIndices(re *regexp.Regexp) []int {
	var indices []int
	seen := make(map[int]bool)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		matches := re.FindStringSubmatch(parts[0])
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

func hasLegacyUserVars() bool {
	return os.Getenv("NUT_USER") != "" || os.Getenv("NUT_PASSWORD") != "" || os.Getenv("NUT_SERVER") != ""
}

func resolveUserPassword(idx int, user *UserConfig) (string, error) {
	if user.SecretName != "" {
		data, err := os.ReadFile(fmt.Sprintf("/run/secrets/%s", user.SecretName))
		if err == nil {
			if pw := strings.TrimSpace(string(data)); pw != "" {
				log.Printf("[mjolnir] user %q password loaded from Docker secret %s", user.Name, user.SecretName)
				return pw, nil
			}
		}
	}

	conventionSecret := fmt.Sprintf("nut-user-%d-password", idx)
	data, err := os.ReadFile(fmt.Sprintf("/run/secrets/%s", conventionSecret))
	if err == nil {
		if pw := strings.TrimSpace(string(data)); pw != "" {
			log.Printf("[mjolnir] user %q password loaded from Docker secret %s", user.Name, conventionSecret)
			return pw, nil
		}
	}

	envKey := fmt.Sprintf("NUT_USER_%d_PASSWORD", idx)
	if pw := os.Getenv(envKey); pw != "" {
		return pw, nil
	}

	return "", fmt.Errorf("no password for user %q (set NUT_USER_%d_PASSWORD, NUT_USER_%d_SECRET_NAME, or mount Docker secret %s)", user.Name, idx, idx, conventionSecret)
}

func resolveLegacyPassword(secretName string) string {
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

func validateUser(user *UserConfig, idx int) error {
	for i, action := range user.Actions {
		upper := strings.ToUpper(action)
		if upper != "SET" && upper != "FSD" {
			return fmt.Errorf("invalid action %q for user %q (NUT_USER_%d_ACTIONS): must be SET or FSD", action, user.Name, idx)
		}
		user.Actions[i] = upper
	}

	if user.Upsmon != "" && user.Upsmon != "primary" && user.Upsmon != "secondary" {
		return fmt.Errorf("invalid upsmon role %q for user %q (NUT_USER_%d_UPSMON): must be primary or secondary", user.Upsmon, user.Name, idx)
	}

	return nil
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, v := range strings.Split(s, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			result = append(result, v)
		}
	}
	return result
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
