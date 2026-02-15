#!/bin/sh
set -e

LANGO_DIR="$HOME/.lango"
mkdir -p "$LANGO_DIR"

# Set up passphrase keyfile from Docker secret.
# The keyfile path (~/.lango/keyfile) is blocked by the agent's filesystem tool.
PASSPHRASE_SECRET="${LANGO_PASSPHRASE_FILE:-/run/secrets/lango_passphrase}"
if [ -f "$PASSPHRASE_SECRET" ]; then
  cp "$PASSPHRASE_SECRET" "$LANGO_DIR/keyfile"
  chmod 600 "$LANGO_DIR/keyfile"
fi

# Import config JSON if present and no profile exists yet.
# The mounted file is copied to /tmp before import so the original
# secret remains untouched. The temp copy is auto-deleted after import.
CONFIG_SECRET="${LANGO_CONFIG_FILE:-/run/secrets/lango_config}"
PROFILE_NAME="${LANGO_PROFILE:-default}"

if [ -f "$CONFIG_SECRET" ] && [ ! -f "$LANGO_DIR/lango.db" ]; then
  echo "Importing config as profile '$PROFILE_NAME'..."
  trap 'rm -f /tmp/lango-import.json' EXIT
  cp "$CONFIG_SECRET" /tmp/lango-import.json
  lango config import /tmp/lango-import.json --profile "$PROFILE_NAME"
  rm -f /tmp/lango-import.json
  trap - EXIT
  echo "Config imported successfully."
fi

exec lango "$@"
