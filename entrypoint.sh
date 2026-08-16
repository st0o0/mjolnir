#!/bin/sh
set -e

NUT_RUN="/run/nut"
NUT_CONF="/etc/nut"
UPSMON_PID=

log() { echo "[mjolnir] $*"; }

# ---------------------------------------------------------------------------
# Signal handling
# ---------------------------------------------------------------------------
cleanup() {
  log "shutting down..."
  upsmon -c stop 2>/dev/null || true
  upsd -c stop 2>/dev/null || true
  upsdrvctl stop 2>/dev/null || true
  log "shutdown complete"
  exit 0
}
trap cleanup TERM INT

# ---------------------------------------------------------------------------
# Password resolution
# ---------------------------------------------------------------------------
if [ -f "/run/secrets/${NUT_SECRET_NAME}" ]; then
  NUT_PASSWORD=$(cat "/run/secrets/${NUT_SECRET_NAME}")
elif [ -z "$NUT_PASSWORD" ]; then
  log "WARNING: no password set — use NUT_PASSWORD or Docker secret '${NUT_SECRET_NAME}'"
  NUT_PASSWORD="changeme"
fi

# ---------------------------------------------------------------------------
# Discover UPS units from NUT_UPS_<n>_NAME env vars
# ---------------------------------------------------------------------------
discover_ups_units() {
  env | grep -oE 'NUT_UPS_[0-9]+_NAME=' | sed 's/_NAME=//' | sort -t_ -k3 -n
}

# ---------------------------------------------------------------------------
# ups.conf
# ---------------------------------------------------------------------------
generate_ups_conf() {
  {
    echo "maxretry = 3"

    for prefix in $(discover_ups_units); do
      eval name=\$${prefix}_NAME
      eval driver=\${${prefix}_DRIVER:-usbhid-ups}
      eval port=\${${prefix}_PORT:-auto}
      eval desc=\${${prefix}_DESC:-UPS}
      eval serial=\$${prefix}_SERIAL
      eval vendorid=\$${prefix}_VENDORID
      eval pollinterval=\$${prefix}_POLLINTERVAL
      eval extra=\$${prefix}_EXTRA
      eval sdorder=\$${prefix}_SDORDER

      echo ""
      echo "[$name]"
      echo "  driver = $driver"
      echo "  port = $port"
      echo "  desc = \"$desc\""
      [ -n "$serial" ]       && echo "  serial = \"$serial\""
      [ -n "$vendorid" ]     && echo "  vendorid = $vendorid"
      [ -n "$pollinterval" ] && echo "  pollinterval = $pollinterval"
      [ -n "$sdorder" ]      && echo "  sdorder = $sdorder"

      if [ -n "$extra" ]; then
        echo "$extra" | tr ',' '\n' | while IFS='=' read -r key val; do
          echo "  $key = $val"
        done
      fi
    done
  } > "${NUT_RUN}/ups.conf"
}

build_ups_conf() {
  if [ -f "${NUT_CONF}/local/ups.conf" ]; then
    log "using mounted ups.conf"
    cp "${NUT_CONF}/local/ups.conf" "${NUT_RUN}/ups.conf"
  else
    log "generating ups.conf from env"
    generate_ups_conf
  fi
}

# ---------------------------------------------------------------------------
# upsd.conf
# ---------------------------------------------------------------------------
build_upsd_conf() {
  if [ -f "${NUT_CONF}/local/upsd.conf" ]; then
    log "using mounted upsd.conf"
    cp "${NUT_CONF}/local/upsd.conf" "${NUT_RUN}/upsd.conf"
  else
    log "generating upsd.conf from env"
    cat > "${NUT_RUN}/upsd.conf" <<EOF
LISTEN ${NUT_LISTEN} 3493
MAXAGE ${NUT_MAXAGE}
EOF
  fi

  # ENV overrides on mounted configs
  if [ "${NUT_MAXAGE}" != "15" ] && [ -f "${NUT_CONF}/local/upsd.conf" ]; then
    sed -i "s/^MAXAGE .*/MAXAGE ${NUT_MAXAGE}/" "${NUT_RUN}/upsd.conf"
  fi
  if [ -f "${NUT_CONF}/local/upsd.conf" ]; then
    if ! grep -q "^LISTEN" "${NUT_RUN}/upsd.conf"; then
      echo "LISTEN ${NUT_LISTEN} 3493" >> "${NUT_RUN}/upsd.conf"
    fi
  fi
}

# ---------------------------------------------------------------------------
# upsd.users
# ---------------------------------------------------------------------------
build_upsd_users() {
  if [ -f "${NUT_CONF}/local/upsd.users" ]; then
    log "using mounted upsd.users"
    cp "${NUT_CONF}/local/upsd.users" "${NUT_RUN}/upsd.users"
  else
    log "generating upsd.users from env"
    cat > "${NUT_RUN}/upsd.users" <<EOF
[${NUT_USER}]
  password = ${NUT_PASSWORD}
  upsmon ${NUT_SERVER}
EOF
  fi
}

# ---------------------------------------------------------------------------
# upsmon.conf
# ---------------------------------------------------------------------------
build_upsmon_conf() {
  if [ -f "${NUT_CONF}/local/upsmon.conf" ]; then
    log "using mounted upsmon.conf"
    cp "${NUT_CONF}/local/upsmon.conf" "${NUT_RUN}/upsmon.conf"
  else
    log "generating upsmon.conf from env"
    {
      for prefix in $(discover_ups_units); do
        eval name=\$${prefix}_NAME
        echo "MONITOR ${name}@localhost 1 ${NUT_USER} ${NUT_PASSWORD} ${NUT_SERVER}"
      done
      echo "RUN_AS_USER nut"
    } > "${NUT_RUN}/upsmon.conf"
  fi
}

# ---------------------------------------------------------------------------
# nut.conf
# ---------------------------------------------------------------------------
build_nut_conf() {
  echo "MODE=netserver" > "${NUT_RUN}/nut.conf"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
log "starting mjolnir"

# Always regenerate configs (no .setup cache)
build_nut_conf
build_ups_conf
build_upsd_conf
build_upsd_users
build_upsmon_conf

# Fix permissions
chown nut:nut "${NUT_RUN}"/*
chmod 640 "${NUT_RUN}"/*

# Ensure state directory
mkdir -p /var/run/nut
chown nut:nut /var/run/nut

# Export config path so NUT reads from /run/nut
export NUT_CONFPATH="${NUT_RUN}"

log "starting drivers"
upsdrvctl -u root start

log "starting upsd"
upsd -u nut

log "starting upsmon"
upsmon -F &
UPSMON_PID=$!

log "ready"
wait $UPSMON_PID
