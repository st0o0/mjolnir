## 1. NUT Protocol Client

- [x] 1.1 Implement `Client` struct with `Dial()` constructor, lazy TCP connect, and `Close()` with `LOGOUT`
- [x] 1.2 Implement NUT protocol line parser: `VAR` lines, `BEGIN`/`END` list blocks, `ERR` responses, `OK` responses
- [x] 1.3 Implement `ListVars(upsName string) (map[string]string, error)` — sends `LIST VAR`, parses response block
- [x] 1.4 Implement `GetVar(upsName, varName string) (string, error)` — sends `GET VAR`, parses single response
- [x] 1.5 Implement `ListUPS() ([]UPS, error)` — sends `LIST UPS`, parses response block
- [x] 1.6 Implement automatic reconnect: detect broken connection on read/write error, reconnect and retry once
- [x] 1.7 Add `sync.Mutex` for concurrent access safety

## 2. Querier Interface

- [x] 2.1 Define `Querier` interface with `ListVars(upsName string) (map[string]string, error)` in `internal/nut`
- [x] 2.2 Verify `Client` satisfies `Querier` (compile-time check)

## 3. Integrate with Collector

- [x] 3.1 Update `health.Collector` to accept `Querier` instead of calling `nut.Query()`
- [x] 3.2 Update `Collector.Run()` to use `querier.ListVars()` instead of `nut.Query()`

## 4. Integrate with Health Server

- [x] 4.1 Update `health.Server` to accept `Querier` in constructor
- [x] 4.2 Update `handleHealthz` to use `querier.ListVars()` instead of `nut.Query()`

## 5. Wire Up in main.go

- [x] 5.1 Create `nut.Client` via `Dial("localhost:3493")` after daemon start
- [x] 5.2 Pass client to `health.NewCollector()` and `health.NewServer()`
- [x] 5.3 Add `client.Close()` to shutdown sequence (before daemon stop)

## 6. Remove Legacy Code

- [x] 6.1 Delete `internal/nut/upsc.go`
- [x] 6.2 Delete `internal/nut/upsc_test.go`
- [x] 6.3 Verify no remaining references to `nut.Query` or `nut.QueryVar`

## 7. Tests

- [x] 7.1 Add mock NUT TCP server test helper (listens on random port, responds to protocol commands)
- [x] 7.2 Test `ListVars` with normal response, unknown UPS error, empty variables
- [x] 7.3 Test `GetVar` with normal response, unknown variable error
- [x] 7.4 Test `ListUPS` with zero, one, and multiple UPS units
- [x] 7.5 Test reconnect behavior: kill mock server mid-connection, verify retry succeeds after restart
- [x] 7.6 Test `Close` on connected and disconnected client
- [x] 7.7 Test `Collector` with mock `Querier` — verify `UpdateMetrics` is called with correct data
- [x] 7.8 Test `handleHealthz` with mock `Querier` — verify JSON response structure
