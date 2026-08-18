## MODIFIED Requirements

### Requirement: Config file permissions

Generated and copied configs MUST have restricted permissions. Ownership MUST be set dynamically by resolving the NUT service user at runtime, not assumed from the writing process.

#### Scenario: Ownership and mode
- **WHEN** configs are written to `/run/nut/`
- **THEN** all files are owned by `nut:nut` with mode `640`

#### Scenario: Ownership set after generation
- **WHEN** `config.Generate()` completes and before NUT daemons start
- **THEN** `FixOwnership` resolves the `nut` user via `os/user.Lookup` and chowns all files in the run directory

#### Scenario: Mounted config ownership
- **WHEN** a config file is copied from `/etc/nut/local/` to `/run/nut/`
- **THEN** the copy is also chowned to `nut:nut` with mode `640`, regardless of the source file's ownership
