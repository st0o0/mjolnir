## Purpose

Define the HTTP health and readiness endpoints (`/healthz`, `/readyz`) and the `healthcheck` CLI subcommand used by Docker HEALTHCHECK.

## Requirements

### Requirement: Health endpoint
The system SHALL expose an HTTP health endpoint at `:9550/healthz` that returns the aggregate health status of all monitored UPS units.

#### Scenario: All UPS units healthy
- **WHEN** `upsc` succeeds for all configured UPS units
- **THEN** `GET /healthz` returns HTTP 200 with JSON body `{"status":"healthy","units":{...}}` where each unit shows its status

#### Scenario: One UPS unit unreachable
- **WHEN** `upsc` fails for one of two configured UPS units
- **THEN** `GET /healthz` returns HTTP 503 with JSON body showing the failed unit's error

#### Scenario: No UPS units configured
- **WHEN** no `NUT_UPS_<n>_NAME` ENV vars are set
- **THEN** `GET /healthz` returns HTTP 503 with `{"status":"unhealthy","error":"no UPS units configured"}`

---

### Requirement: Readiness endpoint
The system SHALL expose `:9550/readyz` that returns HTTP 200 only after all NUT daemons (upsdrvctl, upsd, upsmon) have started successfully.

#### Scenario: Before daemon startup completes
- **WHEN** daemons are still starting
- **THEN** `GET /readyz` returns HTTP 503

#### Scenario: After successful startup
- **WHEN** all daemons have started and the first `upsc` poll succeeds
- **THEN** `GET /readyz` returns HTTP 200

---

### Requirement: Healthcheck CLI subcommand
The binary SHALL support a `healthcheck` subcommand that queries `http://localhost:9550/healthz` and exits with code 0 on HTTP 200 or code 1 otherwise. This subcommand SHALL be used as the Docker HEALTHCHECK command.

#### Scenario: Healthy container
- **WHEN** `/healthz` returns HTTP 200
- **THEN** `mjolnir healthcheck` exits with code 0

#### Scenario: Unhealthy container
- **WHEN** `/healthz` returns HTTP 503
- **THEN** `mjolnir healthcheck` exits with code 1
