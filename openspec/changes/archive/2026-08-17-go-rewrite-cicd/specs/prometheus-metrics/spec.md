## ADDED Requirements

### Requirement: Prometheus metrics endpoint
The system SHALL expose a Prometheus-compatible metrics endpoint at `:9550/metrics` serving UPS telemetry collected by polling `upsc`.

#### Scenario: Metrics available after startup
- **WHEN** the container is running with at least one UPS configured
- **THEN** `GET :9550/metrics` returns HTTP 200 with Prometheus text format containing `mjolnir_ups_*` metrics

#### Scenario: No UPS data available
- **WHEN** the container is running but `upsc` fails for a UPS
- **THEN** the metrics endpoint still returns HTTP 200 but the affected UPS metrics are absent or stale (no crash)

### Requirement: UPS status metric
The system SHALL expose `mjolnir_ups_status` as a gauge labeled with `ups="<name>"`. Value SHALL be `1` for online (OL), `0` for on-battery (OB), and `-1` for unknown/error.

#### Scenario: UPS on line power
- **WHEN** `upsc <name> ups.status` returns `OL`
- **THEN** `mjolnir_ups_status{ups="<name>"}` is `1`

#### Scenario: UPS on battery
- **WHEN** `upsc <name> ups.status` returns `OB`
- **THEN** `mjolnir_ups_status{ups="<name>"}` is `0`

### Requirement: UPS telemetry metrics
The system SHALL expose the following gauges per UPS, labeled with `ups="<name>"`:
- `mjolnir_ups_battery_charge_percent` (from `battery.charge`)
- `mjolnir_ups_load_percent` (from `ups.load`)
- `mjolnir_ups_input_voltage` (from `input.voltage`)
- `mjolnir_ups_output_voltage` (from `output.voltage`)
- `mjolnir_ups_battery_voltage` (from `battery.voltage`)

#### Scenario: All telemetry values present
- **WHEN** `upsc <name>` returns values for all five NUT variables
- **THEN** all five `mjolnir_ups_*` gauges are set for that UPS

#### Scenario: Partial telemetry
- **WHEN** `upsc <name>` returns `battery.charge` but not `input.voltage`
- **THEN** `mjolnir_ups_battery_charge_percent` is set and `mjolnir_ups_input_voltage` is absent (not zero)

### Requirement: Metrics poll interval
The system SHALL poll `upsc` at a fixed interval of 15 seconds per UPS unit. Each poll cycle SHALL update all metrics for all configured UPS units.

#### Scenario: Metrics reflect latest poll
- **WHEN** UPS battery charge changes from 100 to 95 between polls
- **THEN** after the next poll cycle, `mjolnir_ups_battery_charge_percent` reports `95`

### Requirement: Multi-UPS metric labels
Each metric SHALL carry a `ups` label matching the UPS name from `NUT_UPS_<n>_NAME`. Multiple UPS units SHALL produce independent metric series.

#### Scenario: Two UPS units with metrics
- **WHEN** `NUT_UPS_1_NAME=ecoflow` and `NUT_UPS_2_NAME=apc` are configured
- **THEN** metrics include both `mjolnir_ups_status{ups="ecoflow"}` and `mjolnir_ups_status{ups="apc"}`
