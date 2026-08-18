## Purpose

Define the Prometheus metrics endpoint: dynamic UPS telemetry gauges generated from all NUT server variables, exposed at `:9550/metrics` with per-UPS labels for multi-UPS support.

## Requirements

### Requirement: Prometheus metrics endpoint
The system SHALL expose a Prometheus-compatible metrics endpoint at `:9550/metrics` serving UPS telemetry collected by querying the NUT server via the native Go NUT protocol client over TCP.

#### Scenario: Metrics available after startup
- **WHEN** the container is running with at least one UPS configured
- **THEN** `GET :9550/metrics` returns HTTP 200 with Prometheus text format containing `mjolnir_*` metrics

#### Scenario: No UPS data available
- **WHEN** the container is running but the NUT protocol client fails to query a UPS
- **THEN** the metrics endpoint still returns HTTP 200, the affected UPS metrics are absent or stale, and `mjolnir_scrape_error{ups="<name>"}` is `1`

---

### Requirement: Dynamic variable metrics
The system SHALL automatically convert every numeric NUT variable to a Prometheus gauge. The metric name SHALL be `mjolnir_<variable>` where dots and hyphens in the NUT variable name are replaced with underscores. All dynamic metrics carry label `ups="<name>"`.

Non-numeric variables SHALL be handled by the string-metrics capability (enum gauges, timestamp gauges, or info gauges) if registered, or reported via the diagnostics endpoint if unregistered. Non-numeric variables not in the string registry SHALL NOT produce metrics.

#### Scenario: Standard numeric variables
- **WHEN** NUT reports `battery.charge=100`, `input.voltage=230.5`, `battery.voltage.nominal=24.0`
- **THEN** `mjolnir_battery_charge{ups="<name>"}` is `100`, `mjolnir_input_voltage{ups="<name>"}` is `230.5`, `mjolnir_battery_voltage_nominal{ups="<name>"}` is `24`

#### Scenario: Registered string variable
- **WHEN** NUT reports `ups.test.result="OK"` (registered in string-metrics)
- **THEN** no `mjolnir_ups_test_result` dynamic float gauge is created; the variable is handled by the enum gauge handler

#### Scenario: Unregistered string variable
- **WHEN** NUT reports `ups.firmware="FW:2.0"` (registered as info gauge in string-metrics)
- **THEN** no dynamic float gauge is created; the variable is handled by the info gauge handler

#### Scenario: Unknown non-numeric variable
- **WHEN** NUT reports `some.vendor.string="foo"` not registered in any handler
- **THEN** no metric is created; the variable appears in the diagnostics endpoint with `handled_by: null`

#### Scenario: Variable appears on later poll
- **WHEN** a NUT variable not seen on the first poll appears on a subsequent poll
- **THEN** a new gauge is dynamically registered and its value is set

#### Scenario: Partial telemetry
- **WHEN** NUT reports `battery.charge` but not `input.voltage`
- **THEN** `mjolnir_battery_charge` is set and `mjolnir_input_voltage` is absent (not zero)

---

### Requirement: UPS status flags
The system SHALL expose `mjolnir_ups_status{ups="<name>", flag="<FLAG>"}` as a gauge for each known status flag. Value SHALL be `1` if the flag is present in the `ups.status` string, `0` if absent.

Known flags: OL, OB, LB, HB, RB, CHRG, DISCHRG, BYPASS, CAL, OFF, OVER, TRIM, BOOST, FSD.

#### Scenario: UPS on line and charging
- **WHEN** `ups.status` is `"OL CHRG"`
- **THEN** `mjolnir_ups_status{flag="OL"}` is `1`, `mjolnir_ups_status{flag="CHRG"}` is `1`, all other flags are `0`

#### Scenario: UPS on battery, low battery, discharging
- **WHEN** `ups.status` is `"OB LB DISCHRG"`
- **THEN** `mjolnir_ups_status{flag="OB"}` is `1`, `mjolnir_ups_status{flag="LB"}` is `1`, `mjolnir_ups_status{flag="DISCHRG"}` is `1`, all other flags are `0`

---

### Requirement: Device info metric
The system SHALL expose `mjolnir_device_info{ups="<name>", model="...", mfr="...", serial="...", type="..."}` as a gauge with value `1`. Labels are populated from NUT variables `device.model`, `device.mfr`, `device.serial`, `device.type`. Missing variables produce empty label values.

#### Scenario: Full device info
- **WHEN** NUT reports `device.model="Smart-UPS 1500"`, `device.mfr="APC"`, `device.serial="AS123"`, `device.type="ups"`
- **THEN** `mjolnir_device_info{ups="myups", model="Smart-UPS 1500", mfr="APC", serial="AS123", type="ups"}` is `1`

#### Scenario: Partial device info
- **WHEN** NUT reports `device.model="EcoFlow"` but no other `device.*` variables
- **THEN** `mjolnir_device_info{ups="myups", model="EcoFlow", mfr="", serial="", type=""}` is `1`

---

### Requirement: Scrape error metric
The system SHALL expose `mjolnir_scrape_error{ups="<name>"}` as a gauge. Value SHALL be `1` when the most recent `ListVars` call for that UPS failed, `0` when it succeeded.

#### Scenario: Successful scrape
- **WHEN** `ListVars("myups")` succeeds
- **THEN** `mjolnir_scrape_error{ups="myups"}` is `0`

#### Scenario: Failed scrape
- **WHEN** `ListVars("myups")` returns an error
- **THEN** `mjolnir_scrape_error{ups="myups"}` is `1` and previously set metrics remain stale

---

### Requirement: Metrics poll interval
The system SHALL poll the NUT server via the native Go client at a fixed interval of 15 seconds per UPS unit. Each poll cycle SHALL update all metrics for all configured UPS units. The client SHALL reuse a persistent TCP connection across poll cycles.

#### Scenario: Metrics reflect latest poll
- **WHEN** UPS battery charge changes from 100 to 95 between polls
- **THEN** after the next poll cycle, `mjolnir_battery_charge` reports `95`

#### Scenario: Connection loss during polling
- **WHEN** the TCP connection to the NUT server drops mid-poll
- **THEN** the collector logs the error, sets scrape_error to 1, the client reconnects, and the next poll cycle succeeds

---

### Requirement: Multi-UPS metric labels
Each metric SHALL carry a `ups` label matching the UPS name from `NUT_UPS_<n>_NAME`. Multiple UPS units SHALL produce independent metric series.

#### Scenario: Two UPS units with metrics
- **WHEN** `NUT_UPS_1_NAME=ecoflow` and `NUT_UPS_2_NAME=apc` are configured
- **THEN** metrics include both `mjolnir_battery_charge{ups="ecoflow"}` and `mjolnir_battery_charge{ups="apc"}`

---

### Requirement: Excluded variables
The following NUT variable prefixes SHALL NOT produce dynamic gauges (they are handled by special-case metrics or are not useful as standalone gauges):
- `ups.status` — handled by flag metric
- `device.*` — handled by device_info metric
- `driver.*` — driver metadata, not telemetry

String variables registered in the string-metrics capability (enum, timestamp, info handlers) SHALL be excluded from dynamic float gauge creation even if their values happen to be parseable as float.
