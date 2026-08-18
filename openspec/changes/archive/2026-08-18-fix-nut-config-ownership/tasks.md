## 1. Implement FixOwnership

- [x] 1.1 Create `internal/config/ownership.go` with `FixOwnership(runDir string) error` — resolve `nut` UID/GID via `os/user.Lookup`, walk `runDir` and `os.Chown` each file and the directory itself
- [x] 1.2 Revert all `0644` permissions back to `0640` in `internal/config/generator.go` (both `os.WriteFile` calls and `Chmod` in `copyFile`)

## 2. Wire into startup

- [x] 2.1 Call `config.FixOwnership(runDir)` in `cmd/mjolnir/main.go` between `config.Generate()` and `nut.NewManager()`

## 3. Tests

- [x] 3.1 Add unit test for `FixOwnership` — create a temp dir with root-owned files, call `FixOwnership`, verify ownership and mode
- [x] 3.2 Run existing tests to verify no regressions (`go test ./...`)
