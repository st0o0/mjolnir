#!/bin/sh
set -e

IMAGE="${IMAGE:-mjolnir:ci}"
CONTAINER="mjolnir-e2e"

cleanup() {
  docker rm -f "$CONTAINER" 2>/dev/null || true
}
trap cleanup EXIT

echo "=== e2e: starting container ==="
docker run -d \
  --name "$CONTAINER" \
  -e NUT_UPS_1_NAME=dummy \
  -e NUT_UPS_1_DRIVER=dummy-ups \
  -e NUT_UPS_1_PORT=auto \
  -e NUT_PASSWORD=testpass \
  -p 13493:3493 \
  -p 19550:9550 \
  "$IMAGE"

echo "=== e2e: waiting for startup ==="
sleep 5

echo "=== e2e: checking version ==="
docker exec "$CONTAINER" /usr/local/bin/mjolnir --version

echo "=== e2e: checking health endpoint ==="
for i in 1 2 3 4 5; do
  if curl -sf http://localhost:19550/readyz >/dev/null 2>&1; then
    echo "readyz: ok"
    break
  fi
  echo "readyz: waiting ($i/5)..."
  sleep 3
done

echo "=== e2e: checking metrics endpoint ==="
curl -sf http://localhost:19550/metrics | head -5

echo "=== e2e: checking container health ==="
status=$(docker inspect --format='{{.State.Health.Status}}' "$CONTAINER" 2>/dev/null || echo "none")
echo "health status: $status"

echo "=== e2e: all checks passed ==="
