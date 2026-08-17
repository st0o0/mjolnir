## Why

Mjolnir polls UPS data by shelling out to `upsc` every 15 seconds per UPS — spawning an OS process, opening a new TCP connection to `upsd`, parsing stdout text, and discarding both. This works but is unnecessarily heavy: each poll forks a C binary, and the approach is hard to unit-test without the real `upsc` binary present. A native Go NUT protocol client would eliminate process spawning, enable persistent connections with automatic reconnect, improve error handling with typed Go errors, and make the collector fully testable with a mock TCP server.

## What Changes

- **New `internal/nut` client**: Go TCP client speaking the NUT network protocol (port 3493) directly. Supports `LIST UPS`, `LIST VAR`, `GET VAR`, authentication (`USERNAME`/`PASSWORD`), and `LOGOUT`. Maintains a persistent connection with reconnect-on-failure.
- **New protocol parser**: Line-based parser for NUT protocol responses (`BEGIN`/`END` blocks, `VAR` lines, `ERR` responses).
- **Replace `upsc` shell-out**: The `health.Collector` switches from `nut.Query()` (which calls `exec.Command("upsc", ...)`) to the new `nut.Client`.
- **Remove `upsc.go`**: The CLI wrapper becomes dead code once the client is in place.
- **Optional UPS auto-discovery**: With `LIST UPS` the collector can discover UPS names at runtime instead of requiring them from config, enabling a fallback when no explicit names are configured.

## Capabilities

### New Capabilities

- `nut-protocol-client`: Native Go client for the NUT network protocol — connection management, command execution, response parsing, error handling, and authentication.

### Modified Capabilities

- `prometheus-metrics`: Requirements change from "polling `upsc`" to "polling via NUT protocol client". The metrics themselves, their names, labels, and semantics remain identical. The data source changes from CLI to TCP.

## Impact

- **Code**: `internal/nut/upsc.go` removed; new `client.go` and `protocol.go` added. `internal/health/collector.go` updated to accept a `nut.Client` instead of calling `nut.Query()`.
- **Dependencies**: No new external dependencies — uses Go stdlib `net`, `bufio`, `fmt`.
- **Docker image**: `upsc` binary is still installed (it's part of the `nut` Alpine package alongside `upsd`/`upsdrvctl`/`upsmon` which we still need), but mjolnir no longer calls it.
- **API**: No changes to exposed HTTP endpoints, metrics names, or Docker environment variables. Fully backwards compatible.
- **Testing**: NUT protocol client is testable with a mock TCP server in unit tests, significantly improving test coverage for the metrics collection path.
