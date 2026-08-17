# mjolnir

NUT (Network UPS Tools) in Docker with built-in Prometheus exporter. Drop-in replacement for `instantlinux/nut-upsd`.

Runs a Go entrypoint that handles config generation, process supervision, and metrics export on `:9550`.

## Setup

```bash
mkdir -p .secrets
echo "your-password" > .secrets/nut-password
docker compose up -d
```

## Configuration

There are three ways to configure NUT — pick whichever fits your setup.

**Environment variables** are the simplest. Define UPS devices with `NUT_UPS_<n>_*`:

```yaml
environment:
  NUT_UPS_1_NAME: ecoflow
  NUT_UPS_1_DRIVER: usbhid-ups
  NUT_UPS_1_PORT: auto
  NUT_UPS_1_DESC: "EcoFlow Delta"
  NUT_USER: admin   # single-user mode (legacy)
  NUT_MAXAGE: "25"
```

**Mounting config files** gives full control. Mount to `/etc/nut/local/`:

```yaml
volumes:
  - ./configs/ups.conf:/etc/nut/local/ups.conf:ro
  - ./configs/upsd.conf:/etc/nut/local/upsd.conf:ro
  - ./configs/upsd.users:/etc/nut/local/upsd.users:ro
  - ./configs/upsmon.conf:/etc/nut/local/upsmon.conf:ro
```

Templates are in `configs/*.example`.

**Both together** works too — mounted files act as the base, env vars override specific values.

### Multiple UPS devices

Just increment the index:

```yaml
environment:
  NUT_UPS_1_NAME: ecoflow
  NUT_UPS_1_DRIVER: usbhid-ups
  NUT_UPS_1_PORT: auto
  NUT_UPS_2_NAME: server-ups
  NUT_UPS_2_DRIVER: snmp-ups
  NUT_UPS_2_PORT: 192.168.1.100
```

### Multiple users

Define users with `NUT_USER_<n>_*` for fine-grained NUT access control:

```yaml
environment:
  NUT_USER_1_NAME: monitor
  NUT_USER_1_UPSMON: primary
  NUT_USER_2_NAME: admin
  NUT_USER_2_ACTIONS: SET,FSD
  NUT_USER_2_INSTCMDS: ALL
  NUT_USER_3_NAME: remote
  NUT_USER_3_UPSMON: secondary
secrets:
  - nut-user-1-password
  - nut-user-2-password
  - nut-user-3-password
```

When `NUT_USER_<n>_*` vars are set, the legacy `NUT_USER`/`NUT_PASSWORD`/`NUT_SERVER` vars are ignored.

Passwords are resolved per user: Docker secret `NUT_USER_<n>_SECRET_NAME` (or convention `nut-user-<n>-password`) → `NUT_USER_<n>_PASSWORD` env var → error. No default password in multi-user mode.

### Environment variable reference

**UPS devices** (`<n>` = 1, 2, 3, ...):

| Variable | Default | Description |
|----------|---------|-------------|
| `NUT_UPS_<n>_NAME` | `ups` | UPS name |
| `NUT_UPS_<n>_DRIVER` | `usbhid-ups` | NUT driver |
| `NUT_UPS_<n>_PORT` | `auto` | Device path or network address |
| `NUT_UPS_<n>_DESC` | `UPS` | Description |
| `NUT_UPS_<n>_SERIAL` | — | Serial number |
| `NUT_UPS_<n>_VENDORID` | — | USB vendor ID |
| `NUT_UPS_<n>_POLLINTERVAL` | — | Poll interval (seconds) |
| `NUT_UPS_<n>_SDORDER` | — | Shutdown order |
| `NUT_UPS_<n>_EXTRA` | — | Extra driver options (`key=val,key=val`) |

**Users** (`<n>` = 1, 2, 3, ...):

| Variable | Default | Description |
|----------|---------|-------------|
| `NUT_USER_<n>_NAME` | — | Username (required) |
| `NUT_USER_<n>_PASSWORD` | — | Password (fallback if no Docker secret) |
| `NUT_USER_<n>_SECRET_NAME` | `nut-user-<n>-password` | Docker secret name |
| `NUT_USER_<n>_UPSMON` | — | `primary` or `secondary` |
| `NUT_USER_<n>_ACTIONS` | — | `SET`, `FSD`, or `SET,FSD` |
| `NUT_USER_<n>_INSTCMDS` | — | `ALL` or comma-separated commands |

**Legacy single-user** (used when no `NUT_USER_<n>_*` vars are set):

| Variable | Default | Description |
|----------|---------|-------------|
| `NUT_USER` | `admin` | API username |
| `NUT_PASSWORD` | — | Password (prefer Docker secret) |
| `NUT_SECRET_NAME` | `nut-password` | Docker secret name |
| `NUT_SERVER` | `primary` | `primary` or `secondary` |

**General:**

| Variable | Default | Description |
|----------|---------|-------------|
| `NUT_LISTEN` | `0.0.0.0` | Listen address |
| `NUT_MAXAGE` | `15` | Max driver age (seconds) |

## Prometheus metrics

Scrape `:9550/metrics`:

```yaml
scrape_configs:
  - job_name: mjolnir
    static_configs:
      - targets: ['mjolnir:9550']
```

Exported metrics:

| Metric | Type |
|--------|------|
| `mjolnir_ups_status` | 1=online, 0=on-battery, -1=unknown |
| `mjolnir_ups_battery_charge_percent` | Battery charge % |
| `mjolnir_ups_load_percent` | Load % |
| `mjolnir_ups_input_voltage` | Input voltage |
| `mjolnir_ups_output_voltage` | Output voltage |
| `mjolnir_ups_battery_voltage` | Battery voltage |

All metrics carry a `ups="<name>"` label.

## Health checks

| Endpoint | Status |
|----------|--------|
| `GET :9550/healthz` | 200 if all UPS units respond, 503 otherwise |
| `GET :9550/readyz` | 200 after daemons started, 503 during startup |
| `GET :9550/metrics` | Prometheus metrics |

## USB access

Pass the USB bus instead of running privileged:

```yaml
devices:
  - /dev/bus/usb:/dev/bus/usb
```

## Querying UPS status

```bash
docker exec mjolnir upsc ecoflow@localhost
upsc ecoflow@<host-ip>:3493
```

## Listing available drivers

```bash
docker run --rm ghcr.io/st0o0/mjolnir --version
```

Builds for `linux/amd64` and `linux/arm64`. Releases are signed with cosign and include SLSA provenance.
