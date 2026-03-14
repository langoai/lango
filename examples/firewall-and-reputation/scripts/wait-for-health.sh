#!/bin/sh
# wait-for-health.sh <url> [timeout_seconds]
# Waits until the given URL returns HTTP 200.

URL="$1"
TIMEOUT="${2:-60}"
ELAPSED=0

echo "Waiting for $URL to become healthy (timeout: ${TIMEOUT}s)..."
while true; do
  if curl -sf "$URL" >/dev/null 2>&1; then
    echo "$URL is healthy."
    exit 0
  fi

  ELAPSED=$((ELAPSED + 2))
  if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
    echo "ERROR: $URL did not become healthy within ${TIMEOUT}s."
    exit 1
  fi

  sleep 2
done
