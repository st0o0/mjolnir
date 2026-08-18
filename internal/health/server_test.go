package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/st0o0/mjolnir/internal/nut"
)

type mockQuerier struct {
	vars map[string]map[string]string
	err  error
}

func (m *mockQuerier) ListVars(upsName string) (map[string]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	if v, ok := m.vars[upsName]; ok {
		return v, nil
	}
	return nil, &nut.NUTError{Code: "UNKNOWN-UPS"}
}

func TestHealthzNoUPS(t *testing.T) {
	s := NewServer(":0", &mockQuerier{}, nil, NewDiagnosticsStore())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "unhealthy" {
		t.Fatalf("expected status unhealthy, got %s", body["status"])
	}
	if body["error"] != "no UPS units configured" {
		t.Fatalf("expected 'no UPS units configured', got %s", body["error"])
	}
}

func TestHealthzHealthy(t *testing.T) {
	q := &mockQuerier{
		vars: map[string]map[string]string{
			"ecoflow": {"ups.status": "OL", "battery.charge": "100"},
		},
	}
	s := NewServer(":0", q, []string{"ecoflow"}, NewDiagnosticsStore())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "healthy" {
		t.Fatalf("expected status healthy, got %v", body["status"])
	}
}

func TestHealthzUnhealthy(t *testing.T) {
	q := &mockQuerier{
		vars: map[string]map[string]string{},
	}
	s := NewServer(":0", q, []string{"missing"}, NewDiagnosticsStore())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestReadyzNotReady(t *testing.T) {
	s := NewServer(":0", &mockQuerier{}, []string{"ups1"}, NewDiagnosticsStore())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var body map[string]bool
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["ready"] {
		t.Fatal("expected ready=false")
	}
}

func TestReadyzReady(t *testing.T) {
	s := NewServer(":0", &mockQuerier{}, []string{"ups1"}, NewDiagnosticsStore())
	s.SetReady()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]bool
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !body["ready"] {
		t.Fatal("expected ready=true")
	}
}

func TestDiagnosticsEmpty(t *testing.T) {
	ds := NewDiagnosticsStore()
	s := NewServer(":0", &mockQuerier{}, []string{"ups1"}, ds)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/diagnostics", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body []UPSDiagnostics
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body) != 0 {
		t.Fatalf("expected empty array, got %d entries", len(body))
	}
}

func TestDiagnosticsWithData(t *testing.T) {
	ds := NewDiagnosticsStore()
	ds.Update("myups", map[string]string{
		"battery.charge":  "100",
		"ups.test.result": "OK",
		"driver.name":     "usbhid-ups",
		"ups.status":      "OL",
		"device.model":    "Smart-UPS",
		"ups.firmware":    "FW:2.0",
	})

	s := NewServer(":0", &mockQuerier{}, []string{"myups"}, ds)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/diagnostics", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body []UPSDiagnostics
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(body))
	}

	vars := body[0].Variables
	if vars["battery.charge"].Type != "numeric" {
		t.Errorf("battery.charge type: got %s, want numeric", vars["battery.charge"].Type)
	}
	if *vars["battery.charge"].HandledBy != "dynamic_gauge" {
		t.Errorf("battery.charge handled_by: got %s, want dynamic_gauge", *vars["battery.charge"].HandledBy)
	}
	if *vars["ups.test.result"].HandledBy != "enum_gauge" {
		t.Errorf("ups.test.result handled_by: got %s, want enum_gauge", *vars["ups.test.result"].HandledBy)
	}
	if *vars["driver.name"].HandledBy != "skipped_driver" {
		t.Errorf("driver.name handled_by: got %s, want skipped_driver", *vars["driver.name"].HandledBy)
	}
	if *vars["ups.status"].HandledBy != "status_flags" {
		t.Errorf("ups.status handled_by: got %s, want status_flags", *vars["ups.status"].HandledBy)
	}
	if *vars["device.model"].HandledBy != "device_info" {
		t.Errorf("device.model handled_by: got %s, want device_info", *vars["device.model"].HandledBy)
	}
	if *vars["ups.firmware"].HandledBy != "info_gauge" {
		t.Errorf("ups.firmware handled_by: got %s, want info_gauge", *vars["ups.firmware"].HandledBy)
	}
}

func TestDiagnosticsUnhandledVariable(t *testing.T) {
	ds := NewDiagnosticsStore()
	ds.Update("myups", map[string]string{
		"experimental.ups.foo": "bar",
	})

	s := NewServer(":0", &mockQuerier{}, []string{"myups"}, ds)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/diagnostics", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	var body []UPSDiagnostics
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	v := body[0].Variables["experimental.ups.foo"]
	if v.HandledBy != nil {
		t.Errorf("expected null handled_by for unhandled var, got %s", *v.HandledBy)
	}
	if v.Type != "string" {
		t.Errorf("expected type string, got %s", v.Type)
	}
}

func TestDiagnosticsMultipleUPS(t *testing.T) {
	ds := NewDiagnosticsStore()
	ds.Update("ups1", map[string]string{"battery.charge": "100"})
	ds.Update("ups2", map[string]string{"battery.charge": "50"})

	s := NewServer(":0", &mockQuerier{}, []string{"ups1", "ups2"}, ds)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/diagnostics", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	var body []UPSDiagnostics
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(body))
	}
}

func TestDiagnosticsPollFailurePreservesData(t *testing.T) {
	ds := NewDiagnosticsStore()
	ds.Update("myups", map[string]string{"battery.charge": "100"})

	snapshot := ds.Snapshot()
	if len(snapshot) != 1 {
		t.Fatalf("expected 1 entry after update, got %d", len(snapshot))
	}
	if snapshot[0].Variables["battery.charge"].Value != "100" {
		t.Error("expected battery.charge=100")
	}
}
