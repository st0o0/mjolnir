# Design: Dynamic Prometheus Metrics

## Architecture

```
NUT Server ──LIST VAR──▶ Client ──▶ map[string]string (all vars)
                                            │
                                    ┌───────▼────────┐
                                    │  MetricWriter   │
                                    │                 │
                                    ├─────────────────┤
                                    │                 │
                          ┌─────────┤  Classify var   │
                          │         │  by name        │
                          │         └─────────────────┘
                          │
              ┌───────────┼──────────────┬──────────────┐
              ▼           ▼              ▼              ▼
        ┌──────────┐ ┌──────────┐ ┌───────────┐ ┌───────────┐
        │ ups.status│ │ device.* │ │ numeric   │ │ non-num   │
        │ → flags  │ │ → info   │ │ → gauge   │ │ → skip    │
        └──────────┘ └──────────┘ └───────────┘ └───────────┘
              │           │              │
              ▼           ▼              ▼
        ┌──────────┐ ┌──────────┐ ┌───────────┐
        │ Gauge per│ │ Gauge=1  │ │ Gauge per │
        │ flag     │ │ labels:  │ │ variable  │
        │ {flag=X} │ │ model,.. │ │ {ups=X}   │
        └──────────┘ └──────────┘ └───────────┘
```

## Metric Naming

NUT variable names are converted to Prometheus metric names:
- Prefix: `mjolnir_`
- Dots (`.`) → underscores (`_`)
- Hyphens (`-`) → underscores (`_`)

Examples:
| NUT Variable | Prometheus Metric |
|---|---|
| `battery.charge` | `mjolnir_battery_charge` |
| `battery.voltage.nominal` | `mjolnir_battery_voltage_nominal` |
| `input.voltage` | `mjolnir_input_voltage` |
| `ups.load` | `mjolnir_ups_load` |
| `output.frequency` | `mjolnir_output_frequency` |

## Four Metric Categories

### 1. Status Flags (`ups.status`)

NUT's `ups.status` is a space-separated string of flags (e.g. `"OL CHRG"`).

Metric: `mjolnir_ups_status{ups=<name>, flag=<flag>}` — value 1 if active, 0 if not.

Known flags (always emitted, even if not present in status string):

| Flag | Meaning |
|---|---|
| OL | On Line (mains power) |
| OB | On Battery |
| LB | Low Battery |
| HB | High Battery |
| RB | Replace Battery |
| CHRG | Charging |
| DISCHRG | Discharging |
| BYPASS | On Bypass |
| CAL | Calibrating |
| OFF | Offline |
| OVER | Overloaded |
| TRIM | Trimming voltage |
| BOOST | Boosting voltage |
| FSD | Forced Shutdown |

### 2. Device Info (`device.*`)

String-valued device metadata becomes labels on a single info metric:

```
mjolnir_device_info{ups="myups", model="Smart-UPS 1500", mfr="APC", serial="AS123", type="ups"} 1
```

Source variables:
- `device.model` → label `model`
- `device.mfr` → label `mfr`
- `device.serial` → label `serial`
- `device.type` → label `type`

Missing variables produce empty label values.

### 3. Dynamic Numeric Gauges (everything else numeric)

Every NUT variable not handled by categories 1-2 is tested with `strconv.ParseFloat`. If it parses, a gauge `mjolnir_<name_underscored>{ups=<name>}` is set. If it doesn't parse, the variable is silently skipped.

### 4. Scrape Error

`mjolnir_scrape_error{ups=<name>}` — Gauge, value 1 when `ListVars` fails, 0 on success. Reset to 0 on successful poll. Allows alerting on stale metrics.

## Implementation: Dynamic Gauge Registry

Since we don't know all variable names at compile time, we need a dynamic gauge registry.

```go
type MetricWriter struct {
    mu     sync.Mutex
    gauges map[string]*prometheus.GaugeVec  // key: prometheus metric name
    // ...
}
```

On each poll:
1. Iterate all vars from `ListVars`
2. Skip `ups.status` and `device.*` (handled separately)
3. Try `ParseFloat` — skip if not numeric
4. Look up or create `*prometheus.GaugeVec` in the registry
5. Set the value with `ups` label

Gauge creation uses `prometheus.NewGaugeVec` + `prometheus.Register` (not `promauto`, since we register dynamically). On first encounter of a new variable name, the gauge is created and registered. Subsequent polls reuse the existing gauge.

## Collector Changes

The collector needs to:
1. Pass errors to `MetricWriter` (for scrape_error metric)
2. Call `MetricWriter.Write(upsName, vars)` instead of `UpdateMetrics`
3. Call `MetricWriter.SetError(upsName)` on ListVars failure

## Test Strategy

- Unit test `varToMetricName()` conversion
- Unit test status flag parsing
- Unit test device info label extraction
- Unit test dynamic gauge creation and update
- Unit test scrape error metric
- Integration test: feed a realistic `LIST VAR` output, verify all expected metrics appear
