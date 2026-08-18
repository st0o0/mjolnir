## Why

Mjolnir currently drops all non-numeric NUT variables (~30+ standard variables across all drivers). This means alerting-critical data like self-test results (`ups.test.result`), active alarms (`ups.alarm`), charger status (`battery.charger.status`), and transfer reasons (`input.transfer.reason`) are invisible to Prometheus. Users cannot alert on a failed battery self-test or detect a "Replace battery!" alarm — the most operationally valuable signals a UPS can produce.

## What Changes

- **String enum metrics**: Variables with known value sets (e.g., `ups.test.result`, `battery.charger.status`, status fields) are exported as flag-style gauges — one label per known value, active=1/inactive=0. Same pattern as the existing `ups.status` flag gauge.
- **Timestamp metrics**: Date variables (`ups.test.date`, `battery.date`, `battery.mfr.date`) are parsed and exported as Unix-timestamp gauges (`_seconds` suffix). Enables alerts like "last self-test > 14 days ago".
- **Info metrics**: Inventory strings (firmware, battery type, UPS type) are exported as info gauges with the string as a label, value=1. Extends the existing `device_info` pattern.
- **Alarm metric**: `ups.alarm` gets a dedicated boolean gauge (`_active`) plus an info gauge with the alarm text as label, with stale-label cleanup when the alarm changes.
- **Diagnostics endpoint**: `GET /diagnostics` returns JSON showing all NUT variables seen from each UPS with their handling status (which metric handles them, or `null` if unhandled). Helps users discover what their hardware reports without reading NUT docs.

## Capabilities

### New Capabilities
- `string-metrics`: Enum, timestamp, and info gauge export for non-numeric NUT variables
- `diagnostics-endpoint`: HTTP endpoint for NUT variable discovery and handling transparency

### Modified Capabilities
- `prometheus-metrics`: The "Non-numeric variable" scenario changes from "silently skipped" to "handled by string-metrics capability or reported via diagnostics"

## Impact

- **Code**: `internal/health/metrics.go` (new string writers), `internal/health/server.go` (diagnostics endpoint), `internal/health/collector.go` (pass vars to diagnostics store)
- **Metrics**: ~15 new metric names added. All use `mjolnir_` prefix and `ups` label. Enum gauges add a `flag`/`status`/`result` label. No change to existing metrics.
- **Cardinality**: Controlled — enum gauges use fixed known-value sets, info gauges produce 1 series per UPS. No open-ended label values.
- **Dependencies**: None — uses existing `prometheus/client_golang` library
- **Breaking**: None — existing metrics unchanged, new metrics are additive
