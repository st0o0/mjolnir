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
	s := NewServer(":0", &mockQuerier{}, nil)

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
	s := NewServer(":0", q, []string{"ecoflow"})

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
	s := NewServer(":0", q, []string{"missing"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestReadyzNotReady(t *testing.T) {
	s := NewServer(":0", &mockQuerier{}, []string{"ups1"})

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
	s := NewServer(":0", &mockQuerier{}, []string{"ups1"})
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

