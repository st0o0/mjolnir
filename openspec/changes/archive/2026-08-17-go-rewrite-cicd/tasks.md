## 1. Project Scaffolding

- [x] 1.1 Initialize Go module (`go mod init github.com/st0o0/mjolnir`) and create `cmd/mjolnir/main.go` with CLI entrypoint (version flag, healthcheck subcommand, default run mode)
- [x] 1.2 Create `internal/config/`, `internal/nut/`, `internal/health/` package directories with placeholder files
- [x] 1.3 Add Go dependencies: `prometheus/client_golang`, any needed stdlib-only deps

## 2. Config Parsing & Generation

- [x] 2.1 Implement `internal/config/config.go`: ENV var scanning with regex, typed `Config` and `UPSConfig` structs, password resolution (Docker Secrets → ENV → default), non-contiguous numbering support
- [x] 2.2 Implement `internal/config/generator.go`: generate `ups.conf`, `upsd.conf`, `upsd.users`, `upsmon.conf`, `nut.conf` to `/run/nut/`, handle mounted config passthrough from `/etc/nut/local/`, EXTRA field expansion
- [x] 2.3 Write unit tests for config parsing (multi-UPS discovery, non-contiguous numbering, defaults, EXTRA expansion, password resolution chain)
- [x] 2.4 Write unit tests for config generation (verify generated file contents match expected NUT format)

## 3. NUT Daemon Lifecycle

- [x] 3.1 Implement `internal/nut/daemon.go`: start daemons in order (upsdrvctl → upsd → upsmon), process management with `os/exec`, structured error handling on startup failure
- [x] 3.2 Implement signal handling: `os/signal.NotifyContext` for SIGTERM/SIGINT, reverse-order shutdown (upsmon → upsd → upsdrvctl), tolerate already-stopped services
- [x] 3.3 Implement `internal/nut/upsc.go`: `upsc` exec wrapper to query UPS status and variables, parse key=value output into a map
- [x] 3.4 Write unit tests for upsc output parsing

## 4. Health & Metrics

- [x] 4.1 Implement `internal/health/server.go`: HTTP server on `:9550` with `/healthz`, `/readyz`, and `/metrics` endpoints
- [x] 4.2 Implement `/healthz`: query each UPS via upsc, return 200 with per-UPS status JSON on success, 503 on any failure
- [x] 4.3 Implement `/readyz`: return 503 until daemons started and first upsc poll succeeds, then 200
- [x] 4.4 Implement `internal/health/metrics.go`: Prometheus collectors for `mjolnir_ups_status`, `mjolnir_ups_battery_charge_percent`, `mjolnir_ups_load_percent`, `mjolnir_ups_input_voltage`, `mjolnir_ups_output_voltage`, `mjolnir_ups_battery_voltage` — all labeled with `ups`
- [x] 4.5 Implement metrics poll loop: 15s interval, update gauges from upsc output per UPS
- [x] 4.6 Implement `healthcheck` CLI subcommand: HTTP GET to `localhost:9550/healthz`, exit 0 on 200, exit 1 otherwise
- [x] 4.7 Write unit tests for metrics registration and health endpoint responses

## 5. Main Entrypoint Integration

- [x] 5.1 Wire `cmd/mjolnir/main.go`: parse config → generate files → set permissions → start daemons → start HTTP server → block on signal → graceful shutdown
- [x] 5.2 Add logging with `[mjolnir]` prefix matching existing log format
- [ ] 5.3 Integration test: verify full startup sequence in a test with mocked NUT binaries (deferred — covered by e2e)

## 6. Dockerfile

- [x] 6.1 Rewrite Dockerfile as multi-stage: `golang:alpine` build stage (static binary), `alpine` runtime with NUT packages, `COPY --from=build`, version injection via `ARG VERSION=dev`
- [x] 6.2 Update HEALTHCHECK to `CMD ["/usr/local/bin/mjolnir", "healthcheck"]` with `--start-period=45s`
- [x] 6.3 Expose port 9550 in addition to 3493
- [x] 6.4 Update `docker-compose.yml`: add port mapping `9550:9550`, update healthcheck
- [x] 6.5 Remove `entrypoint.sh` (replaced by Go binary)

## 7. CI/CD Workflows

- [x] 7.1 Replace `.github/workflows/build-push.yml` with `ci.yml`: unit, lint, build+smoke, e2e jobs with concurrency groups
- [x] 7.2 Create `.github/workflows/release.yml`: release-please + multi-arch Docker build + GHCR push + cosign signing + SLSA/SBOM attestations
- [x] 7.3 Create `.github/workflows/dev-build.yml`: label-triggered, environment-gated dev image push with PR comment
- [x] 7.4 Create `.github/workflows/security.yml`: Trivy scan on PR (Dockerfile/go.mod/go.sum changes) + weekly cron, SARIF upload
- [x] 7.5 Create `.github/workflows/commitlint.yml`: conventional commit enforcement

## 8. Tooling Configs

- [x] 8.1 Add `commitlint.config.mjs` with relaxed conventional commits ruleset
- [x] 8.2 Add `.golangci.yml` with v2 config (exclusion presets, no project-specific overrides yet)
- [x] 8.3 Add `.hadolint.yaml` (failure-threshold: warning, ignore DL3018)
- [x] 8.4 Add `release-please-config.json` (simple type, bump-minor-pre-major) and `.release-please-manifest.json` (start at `0.1.0`)
- [x] 8.5 Add `.github/dependabot.yml` with gomod, github-actions, docker ecosystems (weekly, grouped)

## 9. E2E Tests

- [x] 9.1 Create `tests/e2e/` directory with a compose-based test: build image, start with mock UPS config, verify NUT port 3493 responds, verify metrics endpoint on 9550, verify healthcheck passes
- [x] 9.2 Create smoke test script for CI: `docker run --rm mjolnir:ci --version` exits 0 with version output (inline in ci.yml build job)

## 10. Documentation

- [x] 10.1 Update `README.md`: document new metrics endpoint (port 9550), health endpoints, Prometheus scrape config example, updated Dockerfile description
- [x] 10.2 Update `.dockerignore` to exclude Go build artifacts and test fixtures
