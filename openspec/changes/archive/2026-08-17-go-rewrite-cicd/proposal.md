## Why

Mjolnir's shell-based entrypoint uses `eval` for dynamic ENV discovery and ad-hoc string manipulation for config generation — patterns that are fragile, hard to test, and impossible to extend safely. Bifrost and Eir already run a battle-tested Go + CI/CD template (unit tests, linting, multi-arch release with cosign/SLSA, Trivy scanning). Rewriting mjolnir in Go unifies the toolchain across all three projects, enables type-safe config parsing, and opens the door to Prometheus metrics and a proper health API — things a shell entrypoint can't provide.

## What Changes

- **Replace `entrypoint.sh` with a Go binary** (`cmd/mjolnir/`) that handles ENV parsing, config generation, NUT daemon lifecycle, and signal handling
- **Add internal packages**: `internal/config/` (ENV parsing, config generation), `internal/nut/` (daemon lifecycle, process management), `internal/health/` (HTTP health endpoint + Prometheus metrics)
- **New multi-stage Dockerfile**: `golang:alpine` build stage producing a static binary, copied to Alpine runtime stage with NUT packages
- **Add Prometheus metrics endpoint** exposing UPS status, battery charge, load, input/output voltage
- **Add HTTP health endpoint** replacing the shell-based `upsc` healthcheck with a structured `/healthz` endpoint
- **Adopt full CI/CD pipeline from Bifrost/Eir**: ci.yml (unit+lint+build+smoke+e2e), release.yml (release-please + cosign + SLSA/SBOM attestations), dev-build.yml, security.yml (Trivy), commitlint.yml
- **Add supporting tooling**: golangci-lint config, hadolint config, commitlint config, Dependabot config, release-please config
- **BREAKING**: Dockerfile base changes from pure Alpine to multi-stage Go build; image internals change but ENV interface and port 3493 remain identical

## Capabilities

### New Capabilities

- `go-binary`: Go binary entrypoint replacing shell — ENV parsing, config generation, daemon lifecycle, signal handling
- `prometheus-metrics`: Prometheus metrics endpoint exposing UPS telemetry (battery, load, voltage, status)
- `health-api`: HTTP health endpoint (`/healthz`) with structured health status

### Modified Capabilities

- `ci-cd`: Pipeline expands from single build-push workflow to full 5-workflow suite (ci, release, dev-build, security, commitlint) with release-please, cosign signing, SLSA provenance, SBOM attestations, and Trivy scanning
- `healthcheck`: Docker HEALTHCHECK switches from shell `upsc` command to binary health subcommand (`/mjolnir healthcheck`)
- `lifecycle`: Process management moves from shell trap/wait to Go os/signal + exec.Cmd with structured error handling

## Impact

- **Dockerfile**: Complete rewrite (multi-stage Go build)
- **docker-compose.yml**: Add metrics port mapping, update healthcheck
- **CI/CD**: Replace single workflow with 5 workflows + supporting configs
- **Dependencies**: New Go module with prometheus/client_golang, NUT client library or upsc exec
- **Backwards compatibility**: ENV interface (NUT_UPS_*, NUT_USER, NUT_PASSWORD, NUT_SECRET_NAME) unchanged; port 3493 unchanged; Docker Secrets unchanged
- **New port**: Metrics/health HTTP server (9550)
