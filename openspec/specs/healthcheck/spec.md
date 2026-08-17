## Purpose

Define the Docker healthcheck strategy: the Go binary's `healthcheck` subcommand queries the HTTP health endpoint, with defined timing, start period, and failure behavior.

## Requirements

### Requirement: Read-only healthcheck

The Docker HEALTHCHECK SHALL execute `/mjolnir healthcheck`, which queries the HTTP health endpoint at `localhost:9550/healthz`. It SHALL exit 0 on HTTP 200 (healthy) and exit 1 on any other response or connection failure.

#### Scenario: Healthy UPS
- **WHEN** the health endpoint returns HTTP 200
- **THEN** the healthcheck subcommand exits with code 0 and Docker marks the container as healthy

#### Scenario: Unhealthy UPS
- **WHEN** the health endpoint returns HTTP 503
- **THEN** the healthcheck subcommand exits with code 1 and Docker marks the container as unhealthy

#### Scenario: Health server not yet started
- **WHEN** the HTTP server has not started yet (during daemon startup)
- **THEN** the healthcheck subcommand exits with code 1 (connection refused)

---

### Requirement: Healthcheck timing

Runs every 30s, 5s timeout, 3 retries before marking unhealthy. Start period of 45s to allow daemon initialization.

#### Scenario: Startup grace period
- **WHEN** the container has been running for less than 45 seconds
- **THEN** healthcheck failures do not count toward the unhealthy threshold

---

### Requirement: Healthcheck scope

The default healthcheck SHALL query only the first configured UPS unit.

#### Scenario: Only first UPS is checked
- **WHEN** multiple UPS units are configured
- **THEN** only `NUT_UPS_1_NAME` is queried -- other units are not checked by the default healthcheck
