## MODIFIED Requirements

### Requirement: ENV-based config generation

The system MUST generate all NUT configuration files from environment variables and write them to `/run/nut/` when no config files are mounted.

#### Scenario: Single UPS with multiple users via ENV only
- **WHEN** `NUT_UPS_1_NAME`, `NUT_USER_1_NAME`, `NUT_USER_2_NAME` are set and no files exist under `/etc/nut/local/`
- **THEN** `ups.conf`, `upsd.conf`, `upsd.users`, `upsmon.conf`, and `nut.conf` are generated in `/run/nut/` with values from ENV

#### Scenario: upsd.users generation with multiple users
- **WHEN** `NUT_USER_1_NAME=monitor`, `NUT_USER_1_UPSMON=primary` and `NUT_USER_2_NAME=admin`, `NUT_USER_2_ACTIONS=SET,FSD`, `NUT_USER_2_INSTCMDS=ALL` are set
- **THEN** `/run/nut/upsd.users` contains two sections with all directives

#### Scenario: upsmon.conf uses primary user
- **WHEN** `NUT_USER_1_NAME=monitor`, `NUT_USER_1_UPSMON=primary` and `NUT_USER_2_NAME=admin` (no upsmon role) are set
- **THEN** MONITOR lines in `/run/nut/upsmon.conf` use `monitor` with their password and `primary` role

#### Scenario: nut.conf is always generated
- **WHEN** the container starts (any mode)
- **THEN** `/run/nut/nut.conf` is written with `MODE=netserver` regardless of mounts

#### Scenario: ENV defaults
- **WHEN** only `NUT_UPS_1_NAME` is set (no user ENV vars)
- **THEN** defaults apply: single user `admin` with default password resolution and `upsmon primary`

---

### Requirement: ENV override on mounted configs

Selected ENV variables SHALL override specific values in mounted config files (merge mode).

#### Scenario: No merge for upsd.users
- **WHEN** `/etc/nut/local/upsd.users` is mounted
- **THEN** all `NUT_USER_<n>_*` ENV vars are ignored for that file — mounted file is copied as-is

#### Scenario: No merge for upsmon.conf
- **WHEN** `/etc/nut/local/upsmon.conf` is mounted
- **THEN** ENV-based MONITOR lines are ignored — mounted file is used as-is

#### Scenario: MAXAGE override
- **WHEN** `/etc/nut/local/upsd.conf` is mounted AND `NUT_MAXAGE` is set to a non-default value (not `15`)
- **THEN** the `MAXAGE` directive in the copied `/run/nut/upsd.conf` is replaced with the ENV value

#### Scenario: LISTEN fallback
- **WHEN** `/etc/nut/local/upsd.conf` is mounted AND it does not contain a `LISTEN` directive
- **THEN** `LISTEN ${NUT_LISTEN} 3493` is appended to `/run/nut/upsd.conf`
