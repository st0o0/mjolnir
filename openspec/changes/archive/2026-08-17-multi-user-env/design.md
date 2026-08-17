## Context

Mjolnir generates NUT config files from ENV vars at container start. The current model supports a single user defined by `NUT_USER`, `NUT_PASSWORD`/`NUT_SECRET_NAME`, and `NUT_SERVER`. NUT's `upsd.users` format supports multiple users with distinct roles (upsmon primary/secondary) and permissions (actions: SET/FSD, instcmds). Users needing multi-user setups today must mount a custom `upsd.users` file, bypassing the ENV-only workflow.

The existing UPS discovery pattern (`NUT_UPS_<n>_*`) already proves that indexed ENV vars work well in this project.

## Goals / Non-Goals

**Goals:**
- Support N users via `NUT_USER_<n>_*` ENV vars with all `upsd.users` directives
- Per-user password resolution via Docker Secrets or ENV fallback
- Backwards compatibility: existing `NUT_USER`/`NUT_PASSWORD`/`NUT_SERVER` continue to work when no indexed user vars are set
- Validation at startup: reject invalid configurations with clear error messages

**Non-Goals:**
- Changing how mounted `upsd.users` files work (passthrough stays as-is)
- Supporting dynamic user changes at runtime (configs are generated at start)
- ACL per UPS unit (NUT doesn't support this in `upsd.users`)

## Decisions

### 1. Indexed ENV pattern (`NUT_USER_<n>_*`)

Follow the established `NUT_UPS_<n>_*` pattern. Discovery scans environment for `NUT_USER_<n>_NAME` keys.

**Why not structured ENV (JSON)?** Indexed vars are consistent with the UPS pattern, simpler to set in docker-compose, and don't require escaping.

### 2. Legacy fallback as implicit User 1

When no `NUT_USER_<n>_NAME` vars are found, the existing `NUT_USER`, `NUT_PASSWORD`/`NUT_SECRET_NAME`, and `NUT_SERVER` vars are treated as an implicit single-user config. This preserves full backwards compatibility — existing docker-compose files continue to work unchanged.

When indexed user vars ARE found, legacy vars are ignored and a log warning is emitted if they're also set.

**Why not coexistence (legacy = user 0, indexed = user 1..N)?** Mixing two naming schemes creates confusion about which password/secret wins. Clean separation is easier to reason about.

### 3. Per-user password resolution

Resolution order per user:
1. Docker Secret at `/run/secrets/<NUT_USER_<n>_SECRET_NAME>` (if var is set)
2. Docker Secret at `/run/secrets/nut-user-<n>-password` (convention-based default)
3. `NUT_USER_<n>_PASSWORD` ENV var
4. Error — no default password in multi-user mode

In legacy fallback mode, the existing resolution chain (NUT_SECRET_NAME → NUT_PASSWORD → "changeme") is preserved.

**Why error instead of default?** Multiple users imply a security-conscious setup. Defaulting to "changeme" for a secondary monitor defeats the purpose.

### 4. Comma-separated values for actions and instcmds

`NUT_USER_<n>_ACTIONS=SET,FSD` and `NUT_USER_<n>_INSTCMDS=ALL` or `NUT_USER_<n>_INSTCMDS=test.panel.start,test.panel.stop`.

Each value is written as a separate directive in `upsd.users` (NUT expects one per line).

**Why commas?** Consistent with existing `NUT_UPS_<n>_EXTRA` parsing. ENV vars can't have multiple values natively, and commas are the established separator in this codebase.

### 5. upsmon.conf uses first primary user

`generateUpsmonConf` finds the first user with `upsmon=primary` and uses their credentials for all MONITOR lines. Startup validation ensures exactly one primary exists (unless no user has a upsmon role at all, which is valid for admin-only setups — but then upsmon won't start, which is a user decision).

### 6. Config struct change

```go
type UserConfig struct {
    Name       string
    Password   string
    SecretName string
    Upsmon     string            // "primary", "secondary", or ""
    Actions    []string          // ["SET"], ["FSD"], ["SET","FSD"], or nil
    Instcmds   []string          // ["ALL"], ["cmd1","cmd2"], or nil
}

type Config struct {
    UPSUnits []UPSConfig
    Users    []UserConfig        // replaces User, Password, SecretName, Server
    Listen   string
    MaxAge   string
}
```

The old `User`, `Password`, `SecretName`, `Server` fields are removed from `Config`. Legacy fallback constructs a single `UserConfig` from the old vars before anything else touches it.

## Risks / Trade-offs

- **[Breaking change for programmatic users]** Code that reads `cfg.User` or `cfg.Password` won't compile. → Mitigation: this is internal; no public API. Only `generator.go` and `main.go` are affected.
- **[Docker Secret naming convention]** `nut-user-<n>-password` is a new convention users must learn. → Mitigation: documented in README; explicit `NUT_USER_<n>_SECRET_NAME` overrides the convention.
- **[ENV explosion for many users]** 5+ users with individual secrets gets verbose. → Mitigation: at that point, mounting `upsd.users` is the right call. The ENV path targets 2-3 users.
- **[Validation complexity]** Invalid combinations (e.g., `actions=SET` with `upsmon=primary`) should be caught early. → Mitigation: validate at `Load()` time with clear error messages.
