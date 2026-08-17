## Context

Mjolnir is a Docker container that configures and runs NUT (Network UPS Tools) daemons via a shell entrypoint. The shell script uses `eval` to discover UPS units from numbered ENV vars, generates NUT config files, and manages daemon lifecycle with `trap`/`wait`. Sibling projects Bifrost and Eir follow an identical Go + CI/CD template: multi-stage Dockerfile (golang:alpine → scratch), release-please, cosign signing, SLSA/SBOM attestations, Trivy scanning, and conventional commits. Mjolnir currently has a single build-and-push workflow with no tests, no linting, no signing, and no release automation.

Key difference from Bifrost/Eir: mjolnir cannot use a `scratch` runtime image because it needs NUT binaries (`upsdrvctl`, `upsd`, `upsmon`, `upsc`) and their dependencies (`libusb`, `net-snmp-libs`). The Go binary replaces the shell entrypoint but orchestrates the same NUT daemons.

## Goals / Non-Goals

**Goals:**
- Replace `entrypoint.sh` with a Go binary that handles config generation, daemon lifecycle, and signal handling
- Add Prometheus metrics endpoint for UPS telemetry
- Add HTTP health endpoint with structured status
- Adopt the full Bifrost/Eir CI/CD template (5 workflows + tooling configs)
- Maintain 100% backwards compatibility with existing ENV var interface and Docker Secrets
- Unify project structure across Bifrost, Eir, and Mjolnir

**Non-Goals:**
- NUT protocol implementation in Go (we exec NUT binaries, not replace them)
- Web UI or dashboard
- Auto-discovery of USB devices beyond what NUT drivers already do
- Multi-container or sidecar architecture
- Notification/alerting system (future change)

## Decisions

### 1. Alpine runtime instead of scratch

**Decision**: Multi-stage build with `golang:alpine` build stage, but `alpine` (not `scratch`) as runtime.

**Rationale**: NUT binaries (`upsdrvctl`, `upsd`, `upsmon`, `upsc`) are C programs with shared library dependencies (`libusb`, `net-snmp-libs`). Statically compiling them is impractical. The Go binary is statically compiled, but the NUT ecosystem requires Alpine's package manager.

**Alternatives considered**:
- scratch + static NUT: Not feasible — NUT links against glibc/musl dynamically
- Distroless: No apk, same problem as scratch
- Keep current single-stage Alpine: Loses the Go build isolation and bloats the image with Go toolchain

**Layout**:
```
FROM golang:alpine AS build
  → static Go binary at /mjolnir

FROM alpine AS runtime
  → apk add nut tini libusb net-snmp-libs
  → COPY --from=build /mjolnir /usr/local/bin/mjolnir
```

### 2. Exec-based NUT daemon management

**Decision**: Use `os/exec` to start and manage NUT daemons as child processes. Query UPS status via `upsc` exec for metrics collection.

**Rationale**: NUT's protocol is undocumented beyond the basics, and there's no maintained Go NUT client library. Exec'ing `upsc` is simple, reliable, and matches what the shell script does today. The Go binary adds structured error handling, proper signal propagation, and concurrent process management.

**Alternatives considered**:
- Native NUT protocol client: Would require implementing RFC-style protocol parsing for minimal gain. Can be added later if `upsc` exec becomes a bottleneck.
- CGo bindings to libnutclient: Breaks static compilation of the Go binary, adds build complexity

### 3. Project layout following cmd/internal convention

**Decision**:
```
cmd/mjolnir/
  main.go              # CLI entry, version flag, healthcheck subcommand
internal/
  config/
    config.go          # ENV parsing, UPS discovery, config struct
    generator.go       # NUT config file generation (ups.conf, upsd.conf, etc.)
    generator_test.go
  nut/
    daemon.go          # Daemon lifecycle (start/stop/signal handling)
    upsc.go            # upsc exec wrapper for status queries
  health/
    server.go          # HTTP server: /healthz + /metrics
    metrics.go         # Prometheus collectors for UPS telemetry
```

**Rationale**: Matches Bifrost (`cmd/bifrost/`, `internal/...`) and Eir (`cmd/eir/`, `internal/...`) exactly. The `internal/` boundary prevents external imports while keeping packages focused.

### 4. Config parsing: structured ENV discovery

**Decision**: Replace `eval`-based shell discovery with `os.Environ()` scan + regex matching. Parse `NUT_UPS_<n>_*` into a typed `[]UPSConfig` slice, sorted by `<n>`.

**Rationale**: Type-safe, testable, no eval injection risk. The regex approach handles non-contiguous numbering naturally.

### 5. Prometheus metrics via upsc polling

**Decision**: Poll `upsc` at a configurable interval (default 15s) and expose metrics on `:9550/metrics`. Metrics include: `mjolnir_ups_status` (gauge, OL=1/OB=0), `mjolnir_ups_battery_charge_percent`, `mjolnir_ups_load_percent`, `mjolnir_ups_input_voltage`, `mjolnir_ups_output_voltage`, `mjolnir_ups_battery_voltage`. All labeled with `ups="<name>"`.

**Rationale**: 15s matches NUT's default poll interval. Separate port from NUT's 3493 keeps protocol boundaries clean. Metric names follow Prometheus naming conventions with `mjolnir_` prefix.

### 6. Health endpoint design

**Decision**: HTTP health server on `:9550` serving:
- `/healthz` — returns 200 if all monitored UPS units respond to `upsc`, 503 otherwise. JSON body with per-UPS status.
- `/metrics` — Prometheus metrics endpoint
- `/readyz` — returns 200 once initial daemon startup completes

The Docker HEALTHCHECK becomes `/mjolnir healthcheck` (a CLI subcommand that queries `/healthz`), matching Bifrost's pattern.

**Alternatives considered**:
- Keep `upsc` shell healthcheck: Works but loses structured status and can't share state with metrics
- Separate health and metrics ports: Unnecessary complexity for a single-container service

### 7. CI/CD: direct adoption of Bifrost template

**Decision**: Copy Bifrost's 5 workflows + tooling configs, adapting only project-specific values (image name, smoke test commands, platform list, annotations).

| Workflow | Adaptation from Bifrost |
|----------|------------------------|
| ci.yml | Smoke test: `docker run --rm mjolnir:ci --version` instead of genkey. E2e: compose-based NUT startup test |
| release.yml | Image name → `ghcr.io/st0o0/mjolnir`, annotations updated, platforms keep `amd64`+`arm64` |
| dev-build.yml | Image name swap only |
| security.yml | Paths add `entrypoint.sh` if kept as fallback, otherwise identical |
| commitlint.yml | Identical |

Supporting configs copied verbatim: `commitlint.config.mjs`, `.hadolint.yaml`, `.golangci.yml` (remove bifrost-specific exclusion rules), `release-please-config.json`, `.release-please-manifest.json` (start at `0.1.0`), `.github/dependabot.yml`.

### 8. Version injection

**Decision**: `ARG VERSION=dev` in Dockerfile, injected via `-X main.version=${VERSION}`. CLI prints version with `--version` flag. Release workflow passes the release-please version.

## Risks / Trade-offs

**[NUT binary compatibility]** → The Go binary depends on NUT binaries being available at expected paths. **Mitigation**: Alpine's `nut` package is stable; pin major version via Dependabot.

**[upsc exec overhead for metrics]** → Forking `upsc` every 15s per UPS adds process overhead vs. a native client. **Mitigation**: Overhead is negligible for 1-4 UPS units. If needed later, implement native NUT protocol client.

**[Alpine image size vs scratch]** → Alpine runtime (~8MB base) is larger than scratch. **Mitigation**: Still under 20MB total, which is small for what it provides. NUT binaries require it.

**[Migration for existing users]** → Image internals change completely. **Mitigation**: ENV interface is unchanged, port 3493 is unchanged, Docker Secrets are unchanged. Only new: port 9550 for metrics (optional).

## Open Questions

- Should we add `arm/v7` platform support like Eir, or keep `amd64`+`arm64` only? (Depends on whether anyone runs NUT on 32-bit ARM)
- Should the metrics poll interval be configurable via ENV, or is 15s a sensible hardcoded default?
- Do we need a graceful drain period on shutdown, or is immediate daemon stop sufficient?
