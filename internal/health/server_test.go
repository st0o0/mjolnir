package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzNoUPS(t *testing.T) {
	s := NewServer(":0", nil)

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

func TestReadyzNotReady(t *testing.T) {
	s := NewServer(":0", []string{"ups1"})

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
	s := NewServer(":0", []string{"ups1"})
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

func TestStatusToFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"OL", 1},
		{"OB", 0},
		{"OL CHRG", 1},
		{"OB DISCHRG", 0},
		{"", -1},
		{"UNKNOWN", -1},
	}

	for _, tt := range tests {
		got := statusToFloat(tt.input)
		if got != tt.want {
			t.Errorf("statusToFloat(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
