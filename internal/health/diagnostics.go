package health

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

type VarDiagnostic struct {
	Value     string  `json:"value"`
	Type      string  `json:"type"`
	HandledBy *string `json:"handled_by"`
}

type UPSDiagnostics struct {
	UPS       string                    `json:"ups"`
	Timestamp time.Time                 `json:"timestamp"`
	Variables map[string]VarDiagnostic  `json:"variables"`
}

type DiagnosticsStore struct {
	mu   sync.RWMutex
	data map[string]UPSDiagnostics
}

func NewDiagnosticsStore() *DiagnosticsStore {
	return &DiagnosticsStore{
		data: make(map[string]UPSDiagnostics),
	}
}

func (d *DiagnosticsStore) Update(upsName string, vars map[string]string) {
	diag := UPSDiagnostics{
		UPS:       upsName,
		Timestamp: time.Now().UTC(),
		Variables: make(map[string]VarDiagnostic, len(vars)),
	}

	for k, v := range vars {
		diag.Variables[k] = VarDiagnostic{
			Value:     v,
			Type:      classifyType(v),
			HandledBy: classifyHandler(k, v),
		}
	}

	d.mu.Lock()
	d.data[upsName] = diag
	d.mu.Unlock()
}

func (d *DiagnosticsStore) Snapshot() []UPSDiagnostics {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]UPSDiagnostics, 0, len(d.data))
	for _, v := range d.data {
		result = append(result, v)
	}
	return result
}

func classifyType(val string) string {
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return "numeric"
	}
	return "string"
}

func classifyHandler(nutVar, val string) *string {
	h := func(s string) *string { return &s }

	if nutVar == "ups.status" {
		return h("status_flags")
	}
	if strings.HasPrefix(nutVar, "device.") {
		return h("device_info")
	}
	if strings.HasPrefix(nutVar, "driver.") {
		return h("skipped_driver")
	}
	if _, ok := enumRegistry[nutVar]; ok {
		return h("enum_gauge")
	}
	if _, ok := timestampRegistry[nutVar]; ok {
		return h("timestamp_gauge")
	}
	if nutVar == "ups.alarm" {
		return h("info_gauge")
	}
	for _, def := range infoRegistry {
		for _, v := range def.nutVars {
			if v == nutVar {
				return h("info_gauge")
			}
		}
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return h("dynamic_gauge")
	}
	return nil
}
