## Purpose

Define how mjolnir discovers and configures multiple UPS units from numbered environment variable prefixes (`NUT_UPS_<n>_*`), and how multi-UPS interacts with mounted config files.

## Requirements

### Requirement: Multi-UPS discovery via ENV prefix

The system MUST discover multiple UPS units from numbered ENV variable prefixes `NUT_UPS_<n>_*`.

#### Scenario: Discovery mechanism
- **WHEN** environment contains `NUT_UPS_1_NAME=ecoflow` and `NUT_UPS_2_NAME=server-ups`
- **THEN** both UPS units are discovered, sorted numerically by `<n>`, and written to `ups.conf` as separate sections

#### Scenario: Only NAME triggers discovery
- **WHEN** `NUT_UPS_3_DRIVER=snmp-ups` is set but `NUT_UPS_3_NAME` is not
- **THEN** UPS unit 3 is not discovered (no section generated)

#### Scenario: Non-contiguous numbering
- **WHEN** `NUT_UPS_1_NAME` and `NUT_UPS_5_NAME` are set (gap at 2-4)
- **THEN** both units are discovered — numbering does not need to be contiguous

---

### Requirement: Per-UPS ENV variables

Each UPS unit MUST support a defined set of configuration variables via its prefix.

#### Scenario: Full UPS configuration
- **WHEN** `NUT_UPS_1_NAME=ecoflow`, `NUT_UPS_1_DRIVER=usbhid-ups`, `NUT_UPS_1_PORT=auto`, `NUT_UPS_1_DESC="EcoFlow Delta"`, `NUT_UPS_1_SERIAL="ABC123"`, `NUT_UPS_1_VENDORID=XXXX`, `NUT_UPS_1_POLLINTERVAL=15`, `NUT_UPS_1_SDORDER=1`
- **THEN** all values appear in the `[ecoflow]` section of `ups.conf`

#### Scenario: Extra driver options
- **WHEN** `NUT_UPS_1_EXTRA="offdelay=30,ondelay=60"`
- **THEN** `ups.conf` contains `offdelay = 30` and `ondelay = 60` as separate lines in the UPS section

---

### Requirement: Multi-UPS in upsmon.conf

The system MUST generate a MONITOR line in `upsmon.conf` for each discovered UPS unit.

#### Scenario: Multiple MONITOR lines
- **WHEN** two UPS units are discovered via ENV
- **THEN** `upsmon.conf` contains one `MONITOR <name>@localhost 1 <user> <password> <server>` line per unit

---

### Requirement: Multi-UPS interaction with mounted configs

Mounted config files MUST take precedence over ENV-based multi-UPS generation.

#### Scenario: Mounted ups.conf replaces all ENV UPS definitions
- **WHEN** `/etc/nut/local/ups.conf` is mounted AND `NUT_UPS_<n>_*` ENV vars are set
- **THEN** the mounted file is used — ENV-based UPS definitions are ignored entirely

#### Scenario: Mounted upsmon.conf replaces generated MONITOR lines
- **WHEN** `/etc/nut/local/upsmon.conf` is mounted
- **THEN** the mounted file is used — no MONITOR lines are generated from discovered UPS units
