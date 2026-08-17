## ADDED Requirements

### Requirement: Go binary entrypoint
The system SHALL provide a statically compiled Go binary (`mjolnir`) as the container entrypoint, replacing `entrypoint.sh`. The binary SHALL handle ENV parsing, NUT config file generation, daemon lifecycle management, and signal handling.

#### Scenario: Binary starts and generates configs
- **WHEN** the container starts with `NUT_UPS_1_NAME=ups` and `NUT_UPS_1_DRIVER=usbhid-ups` set
- **THEN** the binary parses all `NUT_UPS_<n>_*` ENV vars into typed config structs and generates `ups.conf`, `upsd.conf`, `upsd.users`, `upsmon.conf`, and `nut.conf` in `/run/nut/`

#### Scenario: Binary respects mounted config files
- **WHEN** a config file is mounted at `/etc/nut/local/ups.conf`
- **THEN** the binary copies it to `/run/nut/ups.conf` instead of generating from ENV, matching existing behavior

### Requirement: ENV var discovery via structured parsing
The system SHALL discover UPS units by scanning environment variables matching the pattern `NUT_UPS_<n>_NAME` where `<n>` is a positive integer. Discovery MUST support non-contiguous numbering and sort units numerically by `<n>`.

#### Scenario: Non-contiguous UPS numbering
- **WHEN** ENV contains `NUT_UPS_1_NAME=ups1` and `NUT_UPS_5_NAME=ups5` (no 2, 3, 4)
- **THEN** both units are discovered and ordered as `[ups1, ups5]`

#### Scenario: Per-UPS config fields
- **WHEN** `NUT_UPS_1_EXTRA=offdelay=30,ondelay=60` is set
- **THEN** the config generator expands it to separate lines `offdelay = 30` and `ondelay = 60` in `ups.conf`

### Requirement: Version reporting
The binary SHALL support a `--version` flag that prints the build version injected at compile time via `-ldflags`. The default version for untagged builds SHALL be `dev`.

#### Scenario: Version flag output
- **WHEN** the binary is invoked with `--version`
- **THEN** it prints the version string (e.g., `mjolnir 1.2.3`) and exits with code 0

### Requirement: Static compilation
The Go binary SHALL be compiled with `CGO_ENABLED=0`, `-trimpath`, and `-ldflags="-s -w"` to produce a statically linked binary with no debug symbols.

#### Scenario: Binary runs without shared libraries
- **WHEN** the compiled binary is copied to a minimal Alpine image (without Go runtime)
- **THEN** it executes without missing library errors
