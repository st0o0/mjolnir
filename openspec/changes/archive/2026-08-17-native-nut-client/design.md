## Context

Mjolnir currently queries UPS data by shelling out to `upsc` (a C binary from the NUT package). This happens in two places:

1. **`health.Collector`** — polls every 15s per UPS to update Prometheus metrics (`internal/health/collector.go:35`)
2. **`health.Server.handleHealthz`** — queries on each `/healthz` request (`internal/health/server.go:85`)

Both call `nut.Query()` which runs `exec.Command("upsc", "<name>@localhost")`, parses stdout, and returns a `map[string]string`. Each invocation forks a process, opens a new TCP connection to `upsd:3493`, and discards both.

The NUT network protocol is a simple line-based text protocol on TCP port 3493. It is well-suited for a native Go implementation.

## Goals / Non-Goals

**Goals:**
- Replace `upsc` CLI shell-out with a native Go TCP client speaking the NUT protocol
- Maintain a persistent connection with automatic reconnect
- Make the collector and health check testable without NUT binaries
- Keep all existing metrics, endpoints, and behavior identical from the user's perspective

**Non-Goals:**
- Replacing NUT daemons (`upsd`, `upsdrvctl`, `upsmon`) — they stay as C binaries
- Implementing NUT SET/INSTCMD (write commands) — read-only client
- TLS support — `upsd` runs on localhost inside the same container
- UPS auto-discovery via `LIST UPS` replacing config — deferred to a future change (the client supports it, but the collector still uses configured names)

## Decisions

### D1: Client as a struct with interface for testability

The `nut.Client` is a concrete struct, but callers (Collector, Server) accept a `nut.Querier` interface:

```go
type Querier interface {
    ListVars(upsName string) (map[string]string, error)
}
```

**Why over passing `Client` directly:** The health server and collector currently call the package-level `nut.Query()` function. Switching to an interface lets tests inject a mock without a TCP server. The interface is minimal (one method) because that's all both callers need today.

**Alternative considered:** Test with a mock TCP server only. Rejected because it adds complexity to every test and couples tests to protocol details.

### D2: Single shared client instance

One `nut.Client` is created in `main.go` and passed to both the `Collector` and `Server`. The client handles its own connection lifecycle (connect, reconnect, close).

```
main.go
  │
  ├── nutClient := nut.Dial("localhost:3493")
  │
  ├── health.NewCollector(nutClient, upsNames, 15s)
  │
  └── health.NewServer(":9550", nutClient, upsNames)
```

**Why not one client per caller:** The NUT protocol is sequential (one command at a time per connection). Since the collector polls every 15s and healthz is infrequent, contention is minimal. A mutex in the client serializes access. Two connections would waste a file descriptor and complicate lifecycle.

**Alternative considered:** Connection pool. Overkill — `upsd` on localhost with <10 queries/minute doesn't need it.

### D3: Lazy connect with reconnect

The client connects on first use, not at construction time. If the connection drops, the next call reconnects transparently. This avoids startup ordering issues (the client can be created before `upsd` is fully up).

```
Dial("localhost:3493")     →  stores address, no TCP yet
client.ListVars("ecoflow") →  connects, sends LIST VAR, returns result
     ... connection drops ...
client.ListVars("ecoflow") →  detects broken conn, reconnects, retries once
```

**Why lazy:** `upsd` starts asynchronously via `upsdrvctl` → `upsd` → `upsmon`. The Go binary creates the client before starting daemons (or right after). Eager connect would need retry logic at startup; lazy connect handles it naturally.

### D4: Protocol parser as internal functions

The protocol parser lives in the same package (`internal/nut`) as unexported functions. No separate `protocol` package.

```
internal/nut/
  client.go       -- Client struct, Dial, ListVars, GetVar, ListUPS, Close
  client_test.go  -- Tests with mock TCP server
  daemon.go       -- Manager (unchanged)
```

**Why not a separate package:** The NUT protocol is only used by this one client. A separate package would add indirection without reuse. The parser functions (`parseLine`, `parseVarLine`, `readList`) are unexported implementation details.

**Files removed:** `upsc.go` and `upsc_test.go` — dead code after migration.

### D5: Error types

NUT protocol errors (`ERR VAR-NOT-SUPPORTED`, `ERR UNKNOWN-UPS`, etc.) are mapped to a single `NUTError` type:

```go
type NUTError struct {
    Code    string // e.g. "UNKNOWN-UPS", "VAR-NOT-SUPPORTED"
    Message string
}
```

Callers can check `errors.As` for `*NUTError` to distinguish protocol errors from connection errors.

## Risks / Trade-offs

**[Risk] Connection state mismatch** — If `upsd` restarts while the client holds a stale connection, the next read returns EOF.
→ Mitigation: On any read/write error, close the connection and retry once with a fresh connection. If the retry also fails, return the error.

**[Risk] Concurrent access from healthz and collector** — Both use the same client instance.
→ Mitigation: A `sync.Mutex` in the client serializes all protocol exchanges. Healthz requests during a collector poll wait briefly (<10ms for a localhost roundtrip).

**[Risk] Protocol parsing edge cases** — Variable values with quotes, newlines, or unusual characters.
→ Mitigation: The NUT protocol defines values as `"<quoted string>"` with no escape sequences. The parser strips the outer quotes. Test with edge cases (empty values, spaces in values).

**[Trade-off] Still depends on NUT daemon binaries** — We only replace the query path, not the daemon management.
→ Acceptable: The daemons handle USB/SNMP hardware communication which is not worth reimplementing.

## Migration Plan

1. Add `client.go` with `Client`, `Dial`, `ListVars`, `GetVar`, `ListUPS`, `Close`
2. Add `Querier` interface
3. Update `health.Collector` to accept `Querier` instead of calling `nut.Query()`
4. Update `health.Server` to accept `Querier` instead of calling `nut.Query()`
5. Update `main.go` to create the client and wire it through
6. Remove `upsc.go` and `upsc_test.go`
7. Update specs to reflect new data source (done in this change's delta specs)

Rollback: Revert the commit. No data migration, no config changes, no API changes.

## Open Questions

None — the scope is well-defined and the protocol is stable (NUT protocol hasn't changed in years).
