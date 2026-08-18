## Why

Generated and copied NUT config files in `/run/nut/` are owned by `root:root` because mjolnir runs as root. The NUT daemons (`upsd -u nut`, `upsmon`) drop privileges to the `nut` user and cannot read files with mode `0640` owned by root. The current workaround (`0644`) makes all configs world-readable, including `upsd.users` which contains passwords. This violates the requirements in both `security/spec.md` and `configuration/spec.md` which mandate `nut:nut` ownership with mode `640`.

## What Changes

- Add a `FixOwnership` function that resolves the `nut` user's UID/GID at runtime via `os/user.Lookup` and chowns all files in `/run/nut/` to `nut:nut`
- Call `FixOwnership` in `main.go` between `config.Generate()` and `mgr.Start()` — after configs are written, before daemons read them
- Revert file permissions from `0644` back to `0640` in `generator.go` — ownership makes the permissive mode unnecessary

## Capabilities

### New Capabilities

_None — this is a bug fix aligning implementation with existing specs._

### Modified Capabilities

- `configuration`: No requirement change — implementation now matches the existing "Ownership and mode" scenario
- `security`: No requirement change — implementation now matches the existing "Config file permissions" scenario

## Impact

- `internal/config/generator.go`: permissions revert `0644` → `0640`
- `internal/config/ownership.go`: new file (~15 lines), `FixOwnership(runDir string) error`
- `cmd/mjolnir/main.go`: one new call between `Generate()` and `NewManager()`
- No API, ENV var, or user-facing behavior changes
- Requires `nut` user to exist in the container image (guaranteed by Alpine's `nut` package)
