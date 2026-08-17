## Why

Mjolnir currently supports only a single NUT user via `NUT_USER` / `NUT_PASSWORD` / `NUT_SERVER`. NUT's `upsd.users` allows multiple users with distinct roles (primary/secondary monitor, admin with SET/FSD actions, restricted instcmds). Users needing this today must mount a custom `upsd.users` file, breaking the ENV-only workflow that is mjolnir's core value proposition. Supporting multiple users via ENV keeps the "zero config files" promise intact for the most common multi-user scenarios.

## What Changes

- **BREAKING**: The existing `NUT_USER`, `NUT_PASSWORD`, `NUT_SERVER` env vars are replaced by indexed `NUT_USER_<n>_*` vars. When no indexed user vars are found, the legacy vars are used as an implicit User 1 for backwards compatibility.
- New ENV schema: `NUT_USER_<n>_NAME`, `NUT_USER_<n>_PASSWORD`, `NUT_USER_<n>_SECRET_NAME`, `NUT_USER_<n>_UPSMON`, `NUT_USER_<n>_ACTIONS`, `NUT_USER_<n>_INSTCMDS`
- Per-user password resolution via Docker Secrets (convention: `/run/secrets/nut-user-<n>-password` or custom `NUT_USER_<n>_SECRET_NAME`)
- Config generator produces multi-user `upsd.users` with all NUT directives (password, upsmon, actions, instcmds)
- `upsmon.conf` MONITOR lines use the first user with `upsmon=primary`
- Startup validation: at least one user with `upsmon` role must exist; actions/instcmds values are validated

## Capabilities

### New Capabilities
- `multi-user`: ENV-based discovery, configuration, and password resolution for multiple NUT users with distinct roles and permissions

### Modified Capabilities
- `configuration`: Config generation now handles N users instead of one; `generateUpsdUsers` and `generateUpsmonConf` change signatures
- `security`: Password resolution becomes per-user; legacy single-password path becomes a fallback

## Impact

- `internal/config/config.go`: `Config` struct gains `[]UserConfig`, `Load()` gets user discovery logic, `resolvePassword()` becomes per-user
- `internal/config/generator.go`: `generateUpsdUsers()` and `generateUpsmonConf()` iterate over user slice
- `internal/config/config_test.go`, `generator_test.go`: New test cases for multi-user scenarios
- `Dockerfile`: ENV defaults updated to reflect new variable names
- `docker-compose.yml`: Example updated with multi-user configuration
- `configs/upsd.users.example`: Updated to show multi-user example
- Mounted `upsd.users` passthrough is unaffected (still copied as-is)
