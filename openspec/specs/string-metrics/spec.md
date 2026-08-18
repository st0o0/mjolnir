## Purpose

Define Prometheus metrics for NUT string variables that cannot be represented as simple float gauges: enum flag gauges for known value sets, Unix timestamp gauges for date strings, info gauges for inventory labels, and alarm metrics.

## Requirements

### Requirement: Enum gauge metrics for string variables
The system SHALL export NUT string variables with known value sets as flag-style Prometheus gauges. Each known value SHALL be a label value, with the active value set to `1` and all other values set to `0`. All enum gauges carry label `ups="<name>"`.

The following variables and their known values SHALL be supported:

| NUT Variable | Metric Name | Label | Known Values |
|---|---|---|---|
| `ups.test.result` | `mjolnir_ups_test_result` | `result` | `OK`, `Failed`, `InProgress`, `Aborted`, `NoTestInit`, `BATTdetach` |
| `battery.charger.status` | `mjolnir_battery_charger_status` | `status` | `charging`, `discharging`, `floating`, `resting` |
| `ups.beeper.status` | `mjolnir_ups_beeper_status` | `status` | `enabled`, `disabled`, `muted` |
| `input.sensitivity` | `mjolnir_input_sensitivity` | `sensitivity` | `low`, `medium`, `high`, `auto` |
| `input.transfer.reason` | `mjolnir_input_transfer_reason` | `reason` | `noTransfer`, `highLineVoltage`, `brownout`, `selfTest`, `forcedReboot`, `inputFreqOutOfRange`, `inputVoltageOutOfRange` |

The following variables SHALL use the shared status value set (`critical-low`, `warning-low`, `good`, `warning-high`, `critical-high`):

| NUT Variable | Metric Name | Label |
|---|---|---|
| `input.voltage.status` | `mjolnir_input_voltage_status` | `status` |
| `input.current.status` | `mjolnir_input_current_status` | `status` |
| `input.frequency.status` | `mjolnir_input_frequency_status` | `status` |
| `outlet.n.voltage.status` | `mjolnir_outlet_n_voltage_status` | `status` |
| `outlet.n.current.status` | `mjolnir_outlet_n_current_status` | `status` |

#### Scenario: Self-test passed
- **WHEN** NUT reports `ups.test.result="OK"`
- **THEN** `mjolnir_ups_test_result{ups="myups", result="OK"}` is `1` and all other result labels are `0`

#### Scenario: Self-test failed
- **WHEN** NUT reports `ups.test.result="Failed"`
- **THEN** `mjolnir_ups_test_result{ups="myups", result="Failed"}` is `1` and `result="OK"` is `0`

#### Scenario: Battery charging
- **WHEN** NUT reports `battery.charger.status="charging"`
- **THEN** `mjolnir_battery_charger_status{ups="myups", status="charging"}` is `1` and other status labels are `0`

#### Scenario: Input voltage warning
- **WHEN** NUT reports `input.voltage.status="warning-low"`
- **THEN** `mjolnir_input_voltage_status{ups="myups", status="warning-low"}` is `1` and other status labels are `0`

#### Scenario: Unknown enum value
- **WHEN** NUT reports `ups.test.result="SomeNewValue"` not in the known set
- **THEN** all `mjolnir_ups_test_result` labels are `0` (no active flag)

#### Scenario: Variable not reported by UPS
- **WHEN** NUT does not report `ups.test.result` at all
- **THEN** no `mjolnir_ups_test_result` metric is emitted

---

### Requirement: Timestamp gauge metrics for date variables
The system SHALL export NUT date string variables as Prometheus gauges containing Unix timestamps (seconds since epoch). The metric name SHALL use a `_seconds` suffix. The parser SHALL attempt the following formats in order: `2006-01-02 15:04:05`, `2006-01-02`, `2006/01/02`, `01/02/2006`, raw integer (epoch seconds).

| NUT Variable | Metric Name |
|---|---|
| `ups.test.date` | `mjolnir_ups_test_date_seconds` |
| `battery.date` | `mjolnir_battery_date_seconds` |
| `battery.date.maintenance` | `mjolnir_battery_date_maintenance_seconds` |
| `battery.mfr.date` | `mjolnir_battery_mfr_date_seconds` |
| `ups.mfr.date` | `mjolnir_ups_mfr_date_seconds` |

#### Scenario: Standard date format
- **WHEN** NUT reports `ups.test.date="2026-08-15 14:30:00"`
- **THEN** `mjolnir_ups_test_date_seconds{ups="myups"}` is the Unix timestamp `1787068200`

#### Scenario: Date-only format
- **WHEN** NUT reports `battery.date="2024-03-01"`
- **THEN** `mjolnir_battery_date_seconds{ups="myups"}` is the Unix timestamp for `2024-03-01T00:00:00Z`

#### Scenario: Unparseable date
- **WHEN** NUT reports `battery.date="not set"`
- **THEN** no `mjolnir_battery_date_seconds` metric is emitted (absent, not zero)

#### Scenario: Epoch seconds
- **WHEN** NUT reports `ups.test.date="1723729800"`
- **THEN** `mjolnir_ups_test_date_seconds{ups="myups"}` is `1723729800`

---

### Requirement: Info gauge metrics for inventory strings
The system SHALL export inventory-type NUT string variables as Prometheus info gauges with the string value as a label and gauge value `1`. When the value changes between polls, the previous label combination SHALL be set to `0` and the new one to `1`.

| NUT Variable(s) | Metric Name | Labels |
|---|---|---|
| `ups.firmware`, `ups.firmware.aux` | `mjolnir_ups_firmware_info` | `version`, `aux` |
| `battery.type` | `mjolnir_battery_type_info` | `type` |
| `ups.type` | `mjolnir_ups_type_info` | `type` |

#### Scenario: Firmware info
- **WHEN** NUT reports `ups.firmware="925.T2 .I"` and `ups.firmware.aux="08.3"`
- **THEN** `mjolnir_ups_firmware_info{ups="myups", version="925.T2 .I", aux="08.3"}` is `1`

#### Scenario: Firmware version changes after upgrade
- **WHEN** firmware changes from `"925.T2 .I"` to `"926.T1 .I"`
- **THEN** old label combination is set to `0`, new combination is set to `1`

#### Scenario: Battery type
- **WHEN** NUT reports `battery.type="PbAcid"`
- **THEN** `mjolnir_battery_type_info{ups="myups", type="PbAcid"}` is `1`

---

### Requirement: Alarm metrics
The system SHALL export `ups.alarm` as two metrics:
1. `mjolnir_ups_alarm_active{ups="<name>"}` -- gauge, `1` if alarm text is non-empty, `0` otherwise
2. `mjolnir_ups_alarm_info{ups="<name>", alarm="<text>"}` -- info gauge, `1` for the current alarm text. When the alarm text changes, the previous label combination SHALL be set to `0`.

#### Scenario: Active alarm
- **WHEN** NUT reports `ups.alarm="Replace battery!"`
- **THEN** `mjolnir_ups_alarm_active{ups="myups"}` is `1` and `mjolnir_ups_alarm_info{ups="myups", alarm="Replace battery!"}` is `1`

#### Scenario: No alarm
- **WHEN** NUT reports `ups.alarm=""` or does not report `ups.alarm`
- **THEN** `mjolnir_ups_alarm_active{ups="myups"}` is `0`

#### Scenario: Alarm changes
- **WHEN** alarm changes from `"Replace battery!"` to `"Temperature high"`
- **THEN** `mjolnir_ups_alarm_info{alarm="Replace battery!"}` becomes `0` and `mjolnir_ups_alarm_info{alarm="Temperature high"}` becomes `1`

#### Scenario: Alarm clears
- **WHEN** alarm was `"Replace battery!"` and now `ups.alarm` is empty
- **THEN** `mjolnir_ups_alarm_active` is `0` and `mjolnir_ups_alarm_info{alarm="Replace battery!"}` is `0`
