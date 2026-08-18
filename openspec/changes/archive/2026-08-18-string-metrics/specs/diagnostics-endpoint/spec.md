## ADDED Requirements

### Requirement: Diagnostics HTTP endpoint
The system SHALL expose `GET :9550/diagnostics` returning a JSON response that lists all NUT variables from the most recent poll for each UPS, with their current value and handling classification.

Each variable entry SHALL include:
- `value`: the raw string value from NUT
- `type`: `"numeric"` or `"string"`
- `handled_by`: one of `"status_flags"`, `"device_info"`, `"dynamic_gauge"`, `"enum_gauge"`, `"timestamp_gauge"`, `"info_gauge"`, `"skipped_driver"`, or `null` (unhandled)

#### Scenario: Diagnostics with mixed variables
- **WHEN** NUT reports `battery.charge=100`, `ups.test.result=OK`, `driver.name=usbhid-ups`, `ups.status=OL`
- **THEN** `GET /diagnostics` returns JSON:
  ```json
  [
    {
      "ups": "myups",
      "timestamp": "2026-08-18T14:30:00Z",
      "variables": {
        "battery.charge": {"value": "100", "type": "numeric", "handled_by": "dynamic_gauge"},
        "ups.test.result": {"value": "OK", "type": "string", "handled_by": "enum_gauge"},
        "driver.name": {"value": "usbhid-ups", "type": "string", "handled_by": "skipped_driver"},
        "ups.status": {"value": "OL", "type": "string", "handled_by": "status_flags"}
      }
    }
  ]
  ```

#### Scenario: Unhandled variable visible
- **WHEN** NUT reports a variable `experimental.ups.foo="bar"` not in any handler registry
- **THEN** `GET /diagnostics` includes `"experimental.ups.foo": {"value": "bar", "type": "string", "handled_by": null}`

#### Scenario: No data yet
- **WHEN** no successful poll has completed
- **THEN** `GET /diagnostics` returns HTTP 200 with an empty JSON array `[]`

#### Scenario: Multiple UPS units
- **WHEN** two UPS units are configured
- **THEN** `GET /diagnostics` returns a JSON array with one entry per UPS

---

### Requirement: Diagnostics data freshness
The diagnostics endpoint SHALL return data from the most recent successful poll for each UPS. The `timestamp` field SHALL reflect when the poll was executed. Failed polls SHALL NOT clear previously successful diagnostics data.

#### Scenario: Poll failure preserves last data
- **WHEN** the most recent poll for `myups` failed but the previous poll succeeded
- **THEN** `GET /diagnostics` returns the data from the last successful poll with its original timestamp
