## MODIFIED Requirements

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

### Requirement: Excluded variables
The following NUT variable prefixes SHALL NOT produce dynamic gauges (they are handled by special-case metrics or are not useful as standalone gauges):
- `ups.status` — handled by flag metric
- `device.*` — handled by device_info metric
- `driver.*` — driver metadata, not telemetry

String variables registered in the string-metrics capability (enum, timestamp, info handlers) SHALL be excluded from dynamic float gauge creation even if their values happen to be parseable as float.
