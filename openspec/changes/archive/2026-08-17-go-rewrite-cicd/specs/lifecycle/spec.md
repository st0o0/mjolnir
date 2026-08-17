## MODIFIED Requirements

### Requirement: Process startup order
Services start in order: `upsdrvctl start` (as root) → `upsd` (as nut) → `upsmon` (foreground). After daemons start, the HTTP health/metrics server starts on `:9550`. The Go binary's main goroutine blocks on signal wait, keeping the container alive.

#### Scenario: Successful startup sequence
- **WHEN** the container starts with valid UPS configuration
- **THEN** daemons start in order (upsdrvctl → upsd → upsmon), then the HTTP server begins serving on `:9550`

#### Scenario: Driver startup failure
- **WHEN** `upsdrvctl start` fails (e.g., no USB device found)
- **THEN** the binary logs the error and exits with a non-zero code without starting upsd or upsmon

### Requirement: Graceful shutdown via signal handling
SIGTERM/SIGINT SHALL trigger reverse-order shutdown managed by the Go binary: upsmon → upsd → upsdrvctl → HTTP server → exit 0. The Go binary uses `os/signal.NotifyContext` for signal capture. Shutdown tolerates already-stopped services (errors from stopping are logged but do not change the exit code).

#### Scenario: SIGTERM graceful shutdown
- **WHEN** the container receives SIGTERM
- **THEN** the binary stops services in reverse order (upsmon, upsd, upsdrvctl), stops the HTTP server, and exits with code 0

#### Scenario: Already-stopped service during shutdown
- **WHEN** `upsd` has already exited before shutdown begins
- **THEN** the binary logs a warning but continues stopping remaining services and exits 0

### Requirement: tini as PID 1
tini handles signal forwarding and zombie reaping. The Go binary runs as the direct child of tini.

#### Scenario: Signal forwarding through tini
- **WHEN** `docker stop` sends SIGTERM to the container
- **THEN** tini forwards it to the Go binary, which executes graceful shutdown
