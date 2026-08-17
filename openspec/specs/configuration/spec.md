## Purpose

Define how mjolnir generates, mounts, and merges NUT configuration files. The config system supports three modes: ENV-only generation, direct mount passthrough, and a merge of mounted files with ENV overrides. All configs are written to `/run/nut/` and regenerated on every container start.

## Requirements

### Requirement: ENV-based config generation

The system MUST generate all NUT configuration files from environment variables and write them to `/run/nut/` when no config files are mounted.

#### Scenario: Single UPS via ENV only
- **WHEN** `NUT_UPS_1_NAME`, `NUT_UPS_1_DRIVER`, `NUT_UPS_1_PORT` are set and no files exist under `/etc/nut/local/`
- **THEN** `ups.conf`, `upsd.conf`, `upsd.users`, `upsmon.conf`, and `nut.conf` are generated in `/run/nut/` with values from ENV

#### Scenario: Single UPS with multiple users via ENV only
- **WHEN** `NUT_UPS_1_NAME`, `NUT_USER_1_NAME`, `NUT_USER_2_NAME` are set and no files exist under `/etc/nut/local/`
- **THEN** `ups.conf`, `upsd.conf`, `upsd.users`, `upsmon.conf`, and `nut.conf` are generated in `/run/nut/` with values from ENV

#### Scenario: upsd.users generation with multiple users
- **WHEN** `NUT_USER_1_NAME=monitor`, `NUT_USER_1_UPSMON=primary` and `NUT_USER_2_NAME=admin`, `NUT_USER_2_ACTIONS=SET,FSD`, `NUT_USER_2_INSTCMDS=ALL` are set
- **THEN** `/run/nut/upsd.users` contains two sections with all directives

#### Scenario: upsmon.conf uses primary user
- **WHEN** `NUT_USER_1_NAME=monitor`, `NUT_USER_1_UPSMON=primary` and `NUT_USER_2_NAME=admin` (no upsmon role) are set
- **THEN** MONITOR lines in `/run/nut/upsmon.conf` use `monitor` with their password and `primary` role

#### Scenario: ENV defaults
- **WHEN** only `NUT_UPS_1_NAME` is set (no user ENV vars)
- **THEN** defaults apply: single user `admin` with default password resolution and `upsmon primary`

#### Scenario: nut.conf is always generated
- **WHEN** the container starts (any mode)
- **THEN** `/run/nut/nut.conf` is written with `MODE=netserver` regardless of mounts

---

### Requirement: Mounted config passthrough

The system MUST support mounting NUT config files to `/etc/nut/local/` for full control. Mounted files SHALL be copied to `/run/nut/` — originals MUST stay untouched.

#### Scenario: Mount replaces ENV generation
- **WHEN** `/etc/nut/local/ups.conf` exists
- **THEN** it is copied to `/run/nut/ups.conf` and ENV-based UPS definitions (`NUT_UPS_<n>_*`) are ignored for that file

#### Scenario: Each config file is independent
- **WHEN** `/etc/nut/local/upsd.conf` is mounted but `/etc/nut/local/ups.conf` is not
- **THEN** `upsd.conf` uses the mounted version, `ups.conf` is generated from ENV

#### Scenario: Mounted files are read-only
- **WHEN** configs are mounted with `:ro`
- **THEN** originals in `/etc/nut/local/` are never modified — only copies in `/run/nut/` are written

---

### Requirement: ENV override on mounted configs

Selected ENV variables SHALL override specific values in mounted config files (merge mode).

#### Scenario: MAXAGE override
- **WHEN** `/etc/nut/local/upsd.conf` is mounted AND `NUT_MAXAGE` is set to a non-default value (not `15`)
- **THEN** the `MAXAGE` directive in the copied `/run/nut/upsd.conf` is replaced with the ENV value

#### Scenario: LISTEN fallback
- **WHEN** `/etc/nut/local/upsd.conf` is mounted AND it does not contain a `LISTEN` directive
- **THEN** `LISTEN ${NUT_LISTEN} 3493` is appended to `/run/nut/upsd.conf`

#### Scenario: No merge for ups.conf
- **WHEN** `/etc/nut/local/ups.conf` is mounted
- **THEN** ENV-based UPS definitions are fully ignored — no merge occurs

#### Scenario: No merge for upsd.users
- **WHEN** `/etc/nut/local/upsd.users` is mounted
- **THEN** all `NUT_USER_<n>_*` ENV vars are ignored for that file -- mounted file is copied as-is

#### Scenario: No merge for upsmon.conf
- **WHEN** `/etc/nut/local/upsmon.conf` is mounted
- **THEN** ENV-based MONITOR lines are ignored -- mounted file is used as-is

---

### Requirement: Config regeneration on every start

The system MUST NOT cache configuration state. All configs SHALL be regenerated on every container start.

#### Scenario: ENV change takes effect on restart
- **WHEN** a user changes an ENV variable and restarts the container
- **THEN** the new value is reflected in `/run/nut/` without any manual cleanup

#### Scenario: No .setup or cache file
- **WHEN** the container starts
- **THEN** no state file is checked or written to decide whether to regenerate configs

---

### Requirement: Config file permissions

Generated and copied configs MUST have restricted permissions.

#### Scenario: Ownership and mode
- **WHEN** configs are written to `/run/nut/`
- **THEN** all files are owned by `nut:nut` with mode `640`
