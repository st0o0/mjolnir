## Context

Mjolnir's `MetricWriter` in `internal/health/metrics.go` iterates over all NUT variables from `ListVars`, skips `ups.status` (handled as flag gauge), `device.*` (handled as info gauge), and `driver.*` (internal), then attempts `strconv.ParseFloat` on the rest. Anything that fails parsing is silently dropped. This loses ~30+ standard NUT string variables including operationally critical ones like `ups.test.result` and `ups.alarm`.

The existing code already demonstrates two patterns for string-to-metric conversion: flag gauges (`ups.status` → one gauge per flag) and info gauges (`device.*` → single gauge with string labels). This change extends both patterns and adds timestamp parsing.

## Goals / Non-Goals

**Goals:**
- Export all operationally relevant string NUT variables as Prometheus metrics
- Maintain controlled label cardinality (no open-ended label values in enum gauges)
- Provide a discovery mechanism for users to see what their hardware reports
- Follow existing code patterns — no new abstractions beyond what's needed

**Non-Goals:**
- Exporting `driver.*` variables (internal NUT metadata, not UPS telemetry)
- Making the string variable registry configurable via ENV vars (hardcoded is fine)
- Historical tracking of alarm changes (just current state)
- Parsing vendor-specific non-standard variables

## Decisions

### Decision 1: Registry-based string variable dispatch

String variables are handled via a static registry map in `metrics.go` that maps NUT variable names to handler types (`enum`, `timestamp`, `info`). The `writeDynamicGauges` method gains a companion `writeStringVars` method. Variables not in the registry and not parseable as float are recorded for diagnostics but produce no metric.

**Why not generic auto-detection?** A variable's value alone can't reliably distinguish "OK" (enum) from "Replace battery!" (freetext) from "2024-03-01" (date). Explicit registration is safer and keeps cardinality predictable.

### Decision 2: Enum gauges mirror the ups.status pattern

Enum string variables use the same flag-gauge pattern as `ups.status`: one gauge with a discriminator label, active value = 1, inactive values = 0. This means for `ups.test.result`, all known values (`OK`, `Failed`, etc.) are emitted on every scrape — Grafana dashboards can use `== 1` filters.

Enum definitions are grouped by value set to reduce duplication — many variables share the same `critical-low/warning-low/good/warning-high/critical-high` set.

### Decision 3: Timestamp parsing with multiple format support

NUT date variables have no single standard format. Drivers emit `YYYY-MM-DD`, `YYYY/MM/DD`, `MM/DD/YYYY`, epoch seconds, and sometimes freetext like `not set`. The parser tries formats in order; on failure, the metric is absent (not zero) and the raw value appears in diagnostics.

### Decision 4: Info gauges with stale-label cleanup

Info gauges (`ups.firmware`, `battery.type`, `ups.alarm`) set value=1 with the string as a label. When the value changes (e.g., alarm clears), the old label combination is set to 0 and the new one to 1. This requires tracking the previous value per UPS per variable.

For `ups.alarm` specifically: an additional `mjolnir_ups_alarm_active{ups}` gauge (1/0) provides a simple alerting target without needing to match label values.

### Decision 5: Diagnostics as HTTP endpoint, not file

`GET /diagnostics` returns JSON from the latest poll data — no filesystem I/O in the container. The collector stores raw vars alongside its metric writes. The endpoint classifies each variable by its handler (`status_flags`, `device_info`, `dynamic_gauge`, `enum_gauge`, `timestamp_gauge`, `info_gauge`, `skipped_driver`, or `null` for unhandled).

**Why not a file?** The container runs read-only rootfs in production. An HTTP endpoint fits the existing `/healthz`, `/readyz`, `/metrics` pattern and needs no volume mount.

## Risks / Trade-offs

**[Label cardinality from unknown enum values]** → If a NUT driver sends a value not in our known set (e.g., a new `ups.test.result` value), the enum gauge won't track it. Mitigation: the diagnostics endpoint surfaces the raw value so users can report it. The unknown value still appears in the `ups` label's enum gauge with all known values at 0 — detectable via "no active flag" alert.

**[Timestamp parsing failures]** → Unusual date formats will fail silently. Mitigation: diagnostics endpoint shows the raw string. Absent metric (not zero) is a safe default — won't trigger false alerts.

**[Stale info gauge labels accumulating]** → If `ups.alarm` cycles through many distinct messages, old label combinations remain at 0. Mitigation: this is bounded in practice (alarms are rare and repeat). The `alarm_active` boolean gauge is the primary alerting target.

**[Maintenance burden of the string registry]** → New NUT variables require code changes. Mitigation: the diagnostics endpoint makes unknown variables visible, and the registry is a simple map — PRs to extend it are trivial.
