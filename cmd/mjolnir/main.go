package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/st0o0/mjolnir/internal/config"
	"github.com/st0o0/mjolnir/internal/health"
	"github.com/st0o0/mjolnir/internal/nut"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version":
			fmt.Printf("mjolnir %s\n", version)
			return
		case "healthcheck":
			os.Exit(runHealthcheck())
		}
	}

	os.Exit(run())
}

func run() int {
	log.Printf("[mjolnir] starting mjolnir %s", version)

	cfg, err := config.Load()
	if err != nil {
		log.Printf("[mjolnir] config error: %v", err)
		return 1
	}

	runDir := "/run/nut"
	localDir := "/etc/nut/local"

	if err := config.Generate(cfg, runDir, localDir); err != nil {
		log.Printf("[mjolnir] config generation error: %v", err)
		return 1
	}

	upsNames := make([]string, len(cfg.UPSUnits))
	for i, u := range cfg.UPSUnits {
		upsNames[i] = u.Name
	}

	mgr := nut.NewManager(runDir)
	if err := mgr.Start(); err != nil {
		log.Printf("[mjolnir] daemon start error: %v", err)
		return 1
	}

	nutClient := nut.Dial("localhost:3493")

	srv := health.NewServer(":9550", nutClient, upsNames)
	if err := srv.Start(); err != nil {
		log.Printf("[mjolnir] health server error: %v", err)
		mgr.Stop()
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	writer := health.NewMetricWriter(prometheus.DefaultRegisterer)
	collector := health.NewCollector(nutClient, writer, upsNames, 15*time.Second)
	go collector.Run(ctx)

	srv.SetReady()
	log.Printf("[mjolnir] ready")

	select {
	case <-ctx.Done():
	}

	nutClient.Close()
	mgr.Stop()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Stop(shutdownCtx)

	return 0
}

func runHealthcheck() int {
	resp, err := http.Get("http://localhost:9550/healthz")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return 0
	}
	return 1
}
