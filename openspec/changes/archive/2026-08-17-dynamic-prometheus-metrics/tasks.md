# Tasks

- [x] Task 1: Implement MetricWriter with dynamic gauge registry — Replace the static `metrics.go` with a `MetricWriter` struct that maintains a `map[string]*prometheus.GaugeVec` for dynamically registered gauges, provides `varToMetricName()` for naming conversion, provides `Write(upsName, vars)` that classifies and routes variables, registers gauges on first encounter, and is thread-safe. Files: `internal/health/metrics.go`

- [x] Task 2: Implement status flag parsing — In `MetricWriter.Write`, handle `ups.status` specially: parse the space-separated flag string, emit `mjolnir_ups_status{ups, flag}` for each of 14 known flags (1 if present, 0 if absent), remove old `statusToFloat`. Files: `internal/health/metrics.go`

- [x] Task 3: Implement device_info metric — In `MetricWriter.Write`, handle `device.*` variables: extract `device.model`, `device.mfr`, `device.serial`, `device.type` into labels on `mjolnir_device_info` gauge (value 1), skip `device.*` from dynamic gauge generation. Files: `internal/health/metrics.go`

- [x] Task 4: Implement scrape error metric — Add `mjolnir_scrape_error{ups}` gauge to MetricWriter with `SetError(upsName)` (sets to 1) and reset to 0 on successful `Write()`. Files: `internal/health/metrics.go`

- [x] Task 5: Update collector to use MetricWriter — Change `Collector` to hold a `*MetricWriter` instead of calling `UpdateMetrics`. On success call `writer.Write(name, vars)`, on failure call `writer.SetError(name)`. Files: `internal/health/collector.go`

- [x] Task 6: Update server and main to use MetricWriter — Create `MetricWriter` in `main.go` and pass to `Collector`. Update wiring as needed. Files: `cmd/mjolnir/main.go`, `internal/health/server.go`

- [x] Task 7: Write unit tests — `TestVarToMetricName`, `TestStatusFlags`, `TestDeviceInfo`, `TestDynamicGauges`, `TestScrapeError`, `TestWrite` end-to-end with realistic NUT output. Files: `internal/health/metrics_test.go`

- [x] Task 8: Update prometheus-metrics spec — Sync the delta spec to the main spec. Files: `openspec/specs/prometheus-metrics/spec.md`
