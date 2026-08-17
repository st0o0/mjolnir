## MODIFIED Requirements

### Requirement: Password resolution chain

Passwords MUST be resolved from the most secure source available, per user.

#### Scenario: Docker Secret with explicit name (indexed user)
- **WHEN** `NUT_USER_1_SECRET_NAME=my-monitor-secret` is set AND `/run/secrets/my-monitor-secret` exists
- **THEN** user 1's password is read from that secret file

#### Scenario: Docker Secret with convention name (indexed user)
- **WHEN** `NUT_USER_2_SECRET_NAME` is NOT set AND `/run/secrets/nut-user-2-password` exists
- **THEN** user 2's password is read from `/run/secrets/nut-user-2-password`

#### Scenario: ENV variable fallback (indexed user)
- **WHEN** no Docker Secret exists for user 3 AND `NUT_USER_3_PASSWORD=abc` is set
- **THEN** the ENV value is used as user 3's password

#### Scenario: No password in multi-user mode
- **WHEN** indexed user vars are used AND user 2 has no secret and no `NUT_USER_2_PASSWORD`
- **THEN** startup fails with an error naming the user — no default password in multi-user mode

#### Scenario: Legacy single-user fallback preserves defaults
- **WHEN** no `NUT_USER_<n>_NAME` vars exist AND `NUT_USER=admin` is set without password
- **THEN** the legacy resolution chain applies: `NUT_SECRET_NAME` → `NUT_PASSWORD` → `changeme` with warning

#### Scenario: Configurable secret name (legacy)
- **WHEN** `NUT_SECRET_NAME=my-custom-secret` is set in legacy mode
- **THEN** the container looks for `/run/secrets/my-custom-secret`
