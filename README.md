# mjolnir

Modern NUT (Network UPS Tools) Docker image with Prometheus metrics. Clean alternative to `instantlinux/nut-upsd`.

## Features

- **Go binary entrypoint** — type-safe config parsing, structured error handling
- **Prometheus metrics** — UPS telemetry on `:9550/metrics` (battery, load, voltage, status)
- **Health endpoints** — `/healthz`, `/readyz` on `:9550`
- **Direct config mounting** — mount to `/etc/nut/local/`, no cache, regenerated every restart
- **ENV + Mount merge** — mounted configs as base, ENV variables override key values
- **Multi-UPS support** — multiple UPS units in one container via `NUT_UPS_<n>_*` env vars
- **Graceful shutdown** — proper SIGTERM handling, stops upsmon → upsd → drivers
- **Multi-arch** — `linux/amd64` + `linux/arm64` (Raspberry Pi)
- **No privileged mode** — uses `--device` for USB access
- **Signed releases** — cosign keyless signing, SLSA provenance, SBOM attestations

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

## Monitoring

### Prometheus

Metrics are exposed on port `9550`:

```yaml
scrape_configs:
  - job_name: mjolnir
    static_configs:
      - targets: ['mjolnir:9550']
```

Available metrics:

| Metric | Description |
|--------|-------------|
| `mjolnir_ups_status` | 1=online, 0=on-battery, -1=unknown |
| `mjolnir_ups_battery_charge_percent` | Battery charge % |
| `mjolnir_ups_load_percent` | UPS load % |
| `mjolnir_ups_input_voltage` | Input voltage |
| `mjolnir_ups_output_voltage` | Output voltage |
| `mjolnir_ups_battery_voltage` | Battery voltage |

All metrics are labeled with `ups="<name>"`.

### Health Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET :9550/healthz` | Aggregate UPS health — 200 if all UPS units respond, 503 otherwise |
| `GET :9550/readyz` | Readiness — 200 after all daemons started, 503 during startup |
| `GET :9550/metrics` | Prometheus metrics |

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
docker run --rm ghcr.io/st0o0/mjolnir --version
```
