## ADDED Requirements

### Requirement: User discovery via indexed ENV vars

The system SHALL discover NUT users by scanning the environment for `NUT_USER_<n>_NAME` keys, where `<n>` is a positive integer. Users MUST be processed in ascending index order.

#### Scenario: Multiple users defined
- **WHEN** `NUT_USER_1_NAME=monitor`, `NUT_USER_2_NAME=admin`, `NUT_USER_3_NAME=remote` are set
- **THEN** three user configurations are loaded in index order (1, 2, 3)

#### Scenario: Sparse indices
- **WHEN** `NUT_USER_1_NAME=monitor` and `NUT_USER_5_NAME=admin` are set (no 2, 3, 4)
- **THEN** both users are discovered and loaded in order (1, 5)

#### Scenario: No indexed users with legacy vars
- **WHEN** no `NUT_USER_<n>_NAME` vars exist AND `NUT_USER=admin` is set
- **THEN** a single user is created from `NUT_USER`, `NUT_PASSWORD`/`NUT_SECRET_NAME`, and `NUT_SERVER` (backwards compatibility)

#### Scenario: No users at all
- **WHEN** no `NUT_USER_<n>_NAME` vars exist AND `NUT_USER` is not set
- **THEN** a single user is created with the default name `admin` and default password resolution (legacy behavior)

#### Scenario: Indexed users override legacy vars
- **WHEN** both `NUT_USER_1_NAME=monitor` and `NUT_USER=admin` are set
- **THEN** only the indexed user is used; `NUT_USER` is ignored and a warning is logged

---

### Requirement: Per-user NUT directives

Each user SHALL support all `upsd.users` directives: `password`, `upsmon`, `actions`, and `instcmds`.

#### Scenario: User with upsmon role
- **WHEN** `NUT_USER_1_UPSMON=primary` is set
- **THEN** the generated `upsd.users` contains `upsmon primary` under that user's section

#### Scenario: User with actions
- **WHEN** `NUT_USER_2_ACTIONS=SET,FSD` is set
- **THEN** the generated `upsd.users` contains `actions = SET` and `actions = FSD` as separate lines

#### Scenario: User with specific instcmds
- **WHEN** `NUT_USER_2_INSTCMDS=test.panel.start,test.panel.stop` is set
- **THEN** the generated `upsd.users` contains `instcmds = test.panel.start` and `instcmds = test.panel.stop` as separate lines

#### Scenario: User with ALL instcmds
- **WHEN** `NUT_USER_2_INSTCMDS=ALL` is set
- **THEN** the generated `upsd.users` contains `instcmds = ALL`

#### Scenario: User with no optional directives
- **WHEN** `NUT_USER_3_NAME=simple` and `NUT_USER_3_PASSWORD=xxx` are set with no other directives
- **THEN** the generated `upsd.users` section contains only `password`

---

### Requirement: Startup validation

The system SHALL validate the user configuration at startup and fail with a clear error message for invalid configurations.

#### Scenario: Invalid actions value
- **WHEN** `NUT_USER_1_ACTIONS=INVALID` is set
- **THEN** startup fails with an error naming the user and the invalid value

#### Scenario: Valid actions values
- **WHEN** `NUT_USER_1_ACTIONS=SET` or `SET,FSD` or `FSD` is set
- **THEN** validation passes (case-insensitive)

#### Scenario: No users found
- **WHEN** no `NUT_USER_<n>_NAME` and no legacy `NUT_USER` vars are set
- **THEN** the default legacy user (admin) is used — this is not an error

#### Scenario: Duplicate user names
- **WHEN** `NUT_USER_1_NAME=admin` and `NUT_USER_2_NAME=admin` are set
- **THEN** startup fails with an error indicating the duplicate name
