# mjölnir

NUT (Network UPS Tools) in Docker with a built-in Prometheus exporter. Go entrypoint handles config generation from environment variables, NUT daemon supervision, and metrics export.

## Architecture

```
cmd/mjolnir/          Entry point, signal handling, healthcheck subcommand
internal/
  config/             NUT_UPS_<n>_* and NUT_USER_<n>_* env var parsing, Docker secret resolution,
                      NUT config file generation (nut.conf, ups.conf, upsd.conf, upsd.users, upsmon.conf),
                      mounted config file support (/etc/nut/local/), file ownership (chown to nut user)
  nut/                NUT daemon manager (upsdrvctl, upsd, upsmon) and NUT protocol client (TCP)
  health/             HTTP server (:9550) with /healthz, /readyz, /diagnostics, /metrics;
                      Prometheus metric writer (mjolnir_* prefix), periodic NUT variable collector
```

## Build

```bash
go build -o mjolnir ./cmd/mjolnir
go test ./...
go vet ./...
golangci-lint run
```

## Conventions

- Go 1.27+, pure Go, no CGO
- Conventional commits: feat/fix/perf/docs/chore/refactor/test/ci/build/deps
- commitlint enforced, release-please for versioning
- golangci-lint v2, hadolint for Dockerfile
- All env vars prefixed with `NUT_` (UPS config) or `MJOLNIR_` (future app config)
- Logging via `log.Printf` with `[mjolnir]` prefix
- Metric names prefixed with `mjolnir_`
- Uses `prometheus.DefaultRegisterer` (NUT exporter pattern)
- Errors wrapped with `fmt.Errorf` and `%w`
- Config loaded from `os.Getenv` with indexed env var discovery (NUT_UPS_<n>_*, NUT_USER_<n>_*)
- Passwords resolved: Docker secret > env var > error (no default in multi-user mode)

## Deployment

Single container bundling NUT daemons and Go entrypoint. USB device passthrough via `/dev/bus/usb`. Published as `ghcr.io/st0o0/mjolnir`.

## Security-sensitive areas

- `internal/config/` - UPS credentials via env vars and Docker secrets (`/run/secrets/`), config file generation with passwords written to `upsd.users` and `upsmon.conf`
- `internal/config/ownership.go` - runs as root to chown NUT config files to the `nut` user
- `internal/nut/` - process supervision (upsdrvctl, upsd, upsmon) running as root/nut
