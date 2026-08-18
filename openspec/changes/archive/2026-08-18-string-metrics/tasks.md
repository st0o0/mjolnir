## 1. String Variable Registry

- [x] 1.1 Add string variable registry types and data structures in `internal/health/metrics.go` — define handler types (enum, timestamp, info) and the static map of NUT variable name → handler config
- [x] 1.2 Define enum value sets: shared status set (`critical-low`, `warning-low`, `good`, `warning-high`, `critical-high`), `ups.test.result`, `battery.charger.status`, `ups.beeper.status`, `input.sensitivity`, `input.transfer.reason`
- [x] 1.3 Define timestamp variable list: `ups.test.date`, `battery.date`, `battery.date.maintenance`, `battery.mfr.date`, `ups.mfr.date`
- [x] 1.4 Define info gauge variable groups: `ups.firmware` + `ups.firmware.aux`, `battery.type`, `ups.type`

## 2. Enum Gauge Writer

- [x] 2.1 Register enum gauge `prometheus.GaugeVec` metrics in `NewMetricWriter` with appropriate label names (`ups` + discriminator label per variable)
- [x] 2.2 Implement `writeEnumGauges` method — iterate registered enum vars, look up value in NUT vars, set active=1 and all others=0
- [x] 2.3 Add unit tests for enum gauges: known value active, unknown value (all zeros), variable absent (no metric)

## 3. Timestamp Gauge Writer

- [x] 3.1 Implement `parseNUTDate` function — try formats in order: `2006-01-02 15:04:05`, `2006-01-02`, `2006/01/02`, `01/02/2006`, raw epoch integer
- [x] 3.2 Register timestamp gauge metrics in `NewMetricWriter` and implement `writeTimestampGauges` method
- [x] 3.3 Add unit tests for timestamp parsing: each date format, unparseable string (metric absent), epoch seconds

## 4. Info Gauge Writer

- [x] 4.1 Add previous-value tracking map (`map[string]string` keyed by `upsName:varName`) to `MetricWriter` for stale-label cleanup
- [x] 4.2 Register info gauge metrics in `NewMetricWriter` and implement `writeInfoGauges` method with stale-label cleanup
- [x] 4.3 Add unit tests for info gauges: initial value, value change (old=0, new=1), value unchanged (idempotent)

## 5. Alarm Metrics

- [x] 5.1 Register `mjolnir_ups_alarm_active` and `mjolnir_ups_alarm_info` gauges in `NewMetricWriter`
- [x] 5.2 Implement `writeAlarm` method — set active=1/0 based on non-empty, info gauge with stale-label cleanup
- [x] 5.3 Add unit tests for alarm: active alarm, alarm clears, alarm changes text

## 6. Integration into Write Pipeline

- [x] 6.1 Update `MetricWriter.Write` to call `writeEnumGauges`, `writeTimestampGauges`, `writeInfoGauges`, `writeAlarm` alongside existing methods
- [x] 6.2 Update `writeDynamicGauges` to skip variables registered in the string registry (prevent double-handling)
- [x] 6.3 Add integration test: full `Write` call with mixed numeric and string vars, verify all metric types produced

## 7. Diagnostics Endpoint

- [x] 7.1 Add `DiagnosticsStore` struct to `internal/health/` — stores latest raw vars per UPS with timestamp, classifies each variable by handler
- [x] 7.2 Update `Collector` to store raw vars in `DiagnosticsStore` on each successful poll
- [x] 7.3 Add `GET /diagnostics` handler to `internal/health/server.go` returning JSON from `DiagnosticsStore`
- [x] 7.4 Add unit tests for diagnostics: variable classification, multiple UPS, stale data on poll failure

## 8. Documentation

- [x] 8.1 Update README.md metrics table with new string metrics
- [x] 8.2 Add `/diagnostics` to the health checks endpoint table in README.md
