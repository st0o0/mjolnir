## Purpose

Define the security model for mjolnir: password resolution from Docker Secrets, device passthrough without privileged mode, restricted file permissions, and minimal image footprint.

## Requirements

### Requirement: Password resolution chain

Passwords MUST be resolved from the most secure source available.

#### Scenario: Docker Secret (preferred)
- **WHEN** `/run/secrets/${NUT_SECRET_NAME}` exists
- **THEN** its contents are used as `NUT_PASSWORD`

#### Scenario: ENV variable fallback
- **WHEN** no Docker Secret exists AND `NUT_PASSWORD` is set in ENV
- **THEN** the ENV value is used

#### Scenario: Fallback with warning
- **WHEN** no Docker Secret exists AND `NUT_PASSWORD` is empty
- **THEN** password defaults to `changeme` and a WARNING is logged

#### Scenario: Configurable secret name
- **WHEN** `NUT_SECRET_NAME=my-custom-secret` is set
- **THEN** the container looks for `/run/secrets/my-custom-secret`

---

### Requirement: No privileged container mode

The container MUST NOT require `--privileged` to access UPS hardware.

#### Scenario: USB device passthrough
- **WHEN** a USB UPS is connected
- **THEN** access is granted via `--device /dev/bus/usb:/dev/bus/usb` in compose/run

---

### Requirement: Restricted file permissions

Config files in `/run/nut/` MUST have restricted permissions.

#### Scenario: Config file permissions
- **WHEN** configs are written to `/run/nut/`
- **THEN** files are owned by `nut:nut` with mode `640` — not world-readable

---

### Requirement: Minimal attack surface

The image MUST include only the minimal set of packages required to run NUT.

#### Scenario: Minimal packages
- **WHEN** the image is built
- **THEN** only `nut`, `tini`, `libusb`, and `net-snmp-libs` are installed — no shell utilities, compilers, or package managers beyond apk
