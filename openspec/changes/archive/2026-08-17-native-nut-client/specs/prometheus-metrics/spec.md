## MODIFIED Requirements

### Requirement: Prometheus metrics endpoint
The system SHALL expose a Prometheus-compatible metrics endpoint at `:9550/metrics` serving UPS telemetry collected by querying the NUT server via the native Go NUT protocol client over TCP.

#### Scenario: Metrics available after startup
- **WHEN** the container is running with at least one UPS configured
- **THEN** `GET :9550/metrics` returns HTTP 200 with Prometheus text format containing `mjolnir_ups_*` metrics

#### Scenario: No UPS data available
- **WHEN** the container is running but the NUT protocol client fails to query a UPS (connection error or ERR response)
- **THEN** the metrics endpoint still returns HTTP 200 but the affected UPS metrics are absent or stale (no crash)

---

### Requirement: Metrics poll interval
The system SHALL poll the NUT server via the native Go client at a fixed interval of 15 seconds per UPS unit. Each poll cycle SHALL update all metrics for all configured UPS units. The client SHALL reuse a persistent TCP connection across poll cycles.

#### Scenario: Metrics reflect latest poll
- **WHEN** UPS battery charge changes from 100 to 95 between polls
- **THEN** after the next poll cycle, `mjolnir_ups_battery_charge_percent` reports `95`

#### Scenario: Connection loss during polling
- **WHEN** the TCP connection to the NUT server drops mid-poll
- **THEN** the collector logs the error, the client reconnects, and the next poll cycle succeeds
