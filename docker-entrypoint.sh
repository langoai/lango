#!/bin/sh
set -e

LANGO_DIR="$HOME/.lango"
mkdir -p "$LANGO_DIR/skills" "$HOME/bin"

# Verify write permissions on critical directories.
# Named Docker volumes can inherit stale ownership from previous builds.
for dir in "$LANGO_DIR" "$LANGO_DIR/skills" "$HOME/bin"; do
  if [ -d "$dir" ] && ! [ -w "$dir" ]; then
    echo "ERROR: $dir is not writable by $(whoami) (uid=$(id -u))." >&2
    echo "  Hint: remove the volume and recreate it: docker volume rm lango-data" >&2
    exit 1
  fi
done

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
