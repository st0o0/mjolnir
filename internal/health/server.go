package health

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/st0o0/mjolnir/internal/nut"
)

type Server struct {
	httpServer *http.Server
	querier    nut.Querier
	upsNames   []string
	ready      atomic.Bool
}

func NewServer(addr string, querier nut.Querier, upsNames []string) *Server {
	s := &Server{
		querier:  querier,
		upsNames: upsNames,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)
	mux.Handle("/metrics", promhttp.Handler())

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}

	log.Printf("[mjolnir] health server listening on %s", s.httpServer.Addr)

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[mjolnir] health server error: %v", err)
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) SetReady() {
	s.ready.Store(true)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if len(s.upsNames) == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "no UPS units configured",
		})
		return
	}

	type unitStatus struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}

	units := make(map[string]unitStatus)
	healthy := true

	for _, name := range s.upsNames {
		vars, err := s.querier.ListVars(name)
		if err != nil {
			units[name] = unitStatus{Error: err.Error()}
			healthy = false
			continue
		}
		units[name] = unitStatus{Status: vars["ups.status"]}
	}

	status := "healthy"
	code := http.StatusOK
	if !healthy {
		status = "unhealthy"
		code = http.StatusServiceUnavailable
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"units":  units,
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.ready.Load() {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ready": true})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]bool{"ready": false})
	}
}
