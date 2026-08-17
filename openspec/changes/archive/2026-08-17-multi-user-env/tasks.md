## 1. Config struct and user discovery

- [x] 1.1 Add `UserConfig` struct to `internal/config/config.go` with fields: Name, Password, SecretName, Upsmon, Actions, Instcmds
- [x] 1.2 Replace `User`, `Password`, `SecretName`, `Server` fields in `Config` with `Users []UserConfig`
- [x] 1.3 Implement `discoverUserIndices()` scanning for `NUT_USER_<n>_NAME` env vars (analogous to `discoverUPSIndices`)
- [x] 1.4 Implement legacy fallback in `Load()`: when no indexed user vars found, build single `UserConfig` from `NUT_USER`/`NUT_PASSWORD`/`NUT_SECRET_NAME`/`NUT_SERVER`
- [x] 1.5 Log warning when both indexed and legacy user vars are set

## 2. Per-user password resolution

- [x] 2.1 Refactor `resolvePassword` to accept user index and `UserConfig`, resolving: explicit secret name → convention-based secret (`nut-user-<n>-password`) → ENV password → error (multi-user) or default (legacy)
- [x] 2.2 Wire per-user password resolution into `Load()` loop

## 3. Validation

- [x] 3.1 Validate actions values (only SET/FSD allowed, case-insensitive)
- [x] 3.2 Validate no duplicate user names
- [x] 3.3 Add validation tests for all error scenarios (invalid actions, duplicate names, missing password in multi-user mode)

## 4. Config generation

- [x] 4.1 Update `generateUpsdUsers()` to iterate over `[]UserConfig`, writing password, upsmon, actions, and instcmds directives per user
- [x] 4.2 Update `generateUpsmonConf()` to find first user with `upsmon=primary` and use their credentials for MONITOR lines
- [x] 4.3 Update `Generate()` call signature if needed (Config struct change may suffice)

## 5. Tests

- [x] 5.1 Add `config_test.go` tests for multi-user discovery (multiple users, sparse indices, legacy fallback, indexed-overrides-legacy)
- [x] 5.2 Add `generator_test.go` tests for `generateUpsdUsers` with multiple users, actions, instcmds
- [x] 5.3 Add `generator_test.go` tests for `generateUpsmonConf` selecting primary user credentials
- [x] 5.4 Add test for mounted `upsd.users` passthrough still working (no regression)

## 6. Dockerfile and documentation

- [x] 6.1 Update Dockerfile ENV defaults to document new `NUT_USER_<n>_*` variables (keep legacy defaults for backwards compat)
- [x] 6.2 Update `docker-compose.yml` example with multi-user configuration
- [x] 6.3 Update `configs/upsd.users.example` with multi-user example
- [x] 6.4 Update README with multi-user ENV documentation
