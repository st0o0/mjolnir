# mjolnir

Modern NUT (Network UPS Tools) Docker image. Clean alternative to `instantlinux/nut-upsd`.

## Features

- **Direct config mounting** — mount to `/etc/nut/local/`, no cache, regenerated every restart
- **ENV + Mount merge** — mounted configs as base, ENV variables override key values
- **Multi-UPS support** — multiple UPS units in one container via `NUT_UPS_<n>_*` env vars
- **Graceful shutdown** — proper SIGTERM handling via tini, stops upsmon → upsd → drivers
- **Multi-arch** — `linux/amd64` + `linux/arm64` (Raspberry Pi)
- **No privileged mode** — uses `--device` for USB access
- **Current NUT version** — latest from Alpine repos, not pinned to ancient patches

## Quick Start

```bash
# Create password
mkdir -p .secrets
echo "your-password" > .secrets/nut-password

# Start
docker compose up -d
```

## Configuration

### Option 1: Environment Variables (simplest)

```yaml
environment:
  NUT_UPS_1_NAME: ecoflow
  NUT_UPS_1_DRIVER: usbhid-ups
  NUT_UPS_1_PORT: auto
  NUT_UPS_1_DESC: "EcoFlow Delta"
  NUT_USER: admin
  NUT_MAXAGE: "25"
```

### Option 2: Mount Config Files

```yaml
volumes:
  - ./configs/ups.conf:/etc/nut/local/ups.conf:ro
  - ./configs/upsd.conf:/etc/nut/local/upsd.conf:ro
  - ./configs/upsd.users:/etc/nut/local/upsd.users:ro
  - ./configs/upsmon.conf:/etc/nut/local/upsmon.conf:ro
```

See `configs/*.example` for templates.

### Option 3: Both (merge)

Mount configs and set ENV overrides. Mounted configs are used as base, ENV variables like `NUT_MAXAGE` override specific values.

### Multi-UPS

```yaml
environment:
  NUT_UPS_1_NAME: ecoflow
  NUT_UPS_1_DRIVER: usbhid-ups
  NUT_UPS_1_PORT: auto
  NUT_UPS_2_NAME: server-ups
  NUT_UPS_2_DRIVER: snmp-ups
  NUT_UPS_2_PORT: 192.168.1.100
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NUT_UPS_<n>_NAME` | `ups` | UPS name in ups.conf |
| `NUT_UPS_<n>_DRIVER` | `usbhid-ups` | NUT driver |
| `NUT_UPS_<n>_PORT` | `auto` | Device path or network address |
| `NUT_UPS_<n>_DESC` | `UPS` | Description |
| `NUT_UPS_<n>_SERIAL` | — | Device serial number |
| `NUT_UPS_<n>_VENDORID` | — | USB vendor ID |
| `NUT_UPS_<n>_POLLINTERVAL` | — | Poll interval in seconds |
| `NUT_UPS_<n>_SDORDER` | — | Shutdown order |
| `NUT_UPS_<n>_EXTRA` | — | Extra driver options (`key=val,key=val`) |
| `NUT_USER` | `admin` | API username |
| `NUT_PASSWORD` | — | Password (prefer Docker secret) |
| `NUT_SECRET_NAME` | `nut-password` | Docker secret name |
| `NUT_SERVER` | `primary` | `primary` or `secondary` |
| `NUT_LISTEN` | `0.0.0.0` | Listen address |
| `NUT_MAXAGE` | `15` | Max driver age in seconds |

## USB Device Access

Use `--device` instead of `--privileged`:

```yaml
devices:
  - /dev/bus/usb:/dev/bus/usb
```

## Querying UPS Status

```bash
# From the host
docker exec mjolnir upsc ecoflow@localhost

# From another machine on the network
upsc ecoflow@<host-ip>:3493
```

## Available Drivers

```bash
docker run --rm --entrypoint ls ghcr.io/st0o0/mjolnir /usr/lib/nut/
```
