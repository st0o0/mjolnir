# Dynamic Prometheus Metrics

## Problem

mjolnir currently exposes only 6 hardcoded Prometheus metrics from a fixed list of NUT variables. The NUT client already fetches all variables via `LIST VAR` (typically 40-60 per UPS), but discards everything not in the hardcoded list. This means:

- Users with different UPS models miss variables their hardware reports (e.g. `battery.runtime`, `ups.temperature`, `input.frequency`)
- The status metric (`mjolnir_ups_status`) collapses compound states like `OL CHRG` into a single value, losing flag information
- No device identification metric exists for dashboard labeling
- No scrape-error metric to alert on collection failures
- Adding new metrics requires code changes

The `druggeri/nut_exporter:3.3.0` — the established NUT Prometheus exporter — uses a dynamic approach that automatically converts all NUT variables to metrics.

## Proposal

Replace the hardcoded metric definitions with a dynamic system that automatically converts all numeric NUT variables to Prometheus gauges, plus three special-case metrics:

1. **Dynamic variable gauges** — every numeric NUT variable becomes a `mjolnir_<var_with_underscores>` gauge
2. **Flag-based status metric** — `ups.status` is split into per-flag series (OL, OB, LB, CHRG, etc.)
3. **Device info metric** — `device.*` string variables become labels on a `mjolnir_device_info` gauge
4. **Scrape error metric** — `mjolnir_scrape_error` signals collection failures

## Scope

- Replace `internal/health/metrics.go` with dynamic metric generation
- Update `internal/health/collector.go` to pass scrape errors to metrics
- Update prometheus-metrics spec to reflect the new behavior
- Update tests

## Out of Scope

- Filtering/allowlist configuration (can be added later)
- Custom metric naming overrides
- String-to-boolean regex matching (non-numeric, non-device, non-status vars are simply skipped)

## Breaking Changes

- `mjolnir_ups_status` changes from single-value (1/0/-1) to flag-based series with `flag` label
- Metric names change from `mjolnir_ups_battery_charge_percent` to `mjolnir_battery_charge` (dots→underscores, no `ups_` infix, matching NUT variable names directly)
- Pre-1.0, so acceptable
