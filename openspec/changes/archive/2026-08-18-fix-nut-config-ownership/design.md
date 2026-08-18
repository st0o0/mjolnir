## Context

mjolnir runs as root inside the container (required for USB access via `upsdrvctl`). It generates NUT config files into `/run/nut/` using `os.WriteFile`, which makes them owned by `root:root`. The NUT daemons `upsd` and `upsmon` drop privileges to the `nut` user and need to read these files. The current workaround uses mode `0644` (world-readable), but `upsd.users` contains plaintext passwords and should not be world-readable.

The Dockerfile already runs `chown -R nut:nut /run/nut` at build time, but `config.Generate()` overwrites those files at every container start as root.

## Goals / Non-Goals

**Goals:**
- Config files owned by `nut:nut` with mode `0640` after generation, before daemons start
- Dynamic UID/GID resolution — no hardcoded numeric IDs
- Works for both generated configs and configs copied from `/etc/nut/local/`

**Non-Goals:**
- Supporting `user: nut` in docker-compose — `upsdrvctl` requires root for USB, privilege drop is handled by the NUT daemons themselves
- Changing the NUT daemon startup flags or privilege model
- Making the run directory or permission mode configurable via ENV

## Decisions

### Decision: `os/user.Lookup` for UID/GID resolution

Resolve the `nut` user's UID and GID at runtime via Go's `os/user.Lookup("nut")` instead of hardcoding numeric IDs.

**Why over hardcoded IDs**: Alpine's `nut` package assigns UID/GID dynamically. While it's currently consistent (typically 84), relying on a magic number is fragile across Alpine versions and custom images.

**Why over `os.Exec("chown")`**: Pure Go, no shell dependency, testable, works cross-platform for CI.

### Decision: Separate `FixOwnership` function in `config` package

A dedicated `FixOwnership(runDir string) error` function rather than inlining chown logic into `Generate()` or each `WriteFile` call.

**Why not inline in Generate**: Separation of concerns — `Generate` writes content, `FixOwnership` fixes permissions. Easier to test independently. Also covers the `copyFile` path without modifying it.

**Why not chown per-file during write**: Would require passing UID/GID through every generator function. A single walk after all files are written is simpler and guarantees no file is missed.

### Decision: Revert to `0640` permissions

With correct `nut:nut` ownership, mode `0640` (owner read/write, group read) is sufficient. The `nut` user is the owner and can read all files. No world-readable bit needed.

## Risks / Trade-offs

- **`nut` user doesn't exist** → `user.Lookup` returns an error. Mitigation: this only happens if someone builds a custom image without the `nut` package. The error message will be clear. The Dockerfile's `apk add nut` guarantees the user exists.
- **UID/GID parsing** → `user.Lookup` returns strings. `strconv.Atoi` could fail on non-numeric IDs. Mitigation: Alpine uses numeric IDs in `/etc/passwd`; this is not a realistic failure mode on Linux.
- **Race between Generate and FixOwnership** → No risk: both run sequentially in `main.go` before any daemon starts.

## Open Questions

None — the design is straightforward and all decisions are settled.
