#!/bin/sh
set -e

LANGO_DIR="$HOME/.lango"
mkdir -p "$LANGO_DIR"

# ── Set up passphrase keyfile ──
PASSPHRASE_SECRET="${LANGO_PASSPHRASE_FILE:-/run/secrets/lango_passphrase}"
if [ -f "$PASSPHRASE_SECRET" ]; then
  cp "$PASSPHRASE_SECRET" "$LANGO_DIR/keyfile"
  chmod 600 "$LANGO_DIR/keyfile"
fi

# ── Import config ──
CONFIG_SECRET="${LANGO_CONFIG_FILE:-/run/secrets/lango_config}"
PROFILE_NAME="${LANGO_PROFILE:-default}"

if [ -f "$CONFIG_SECRET" ] && [ ! -f "$LANGO_DIR/lango.db" ]; then
  echo "[$AGENT_NAME] Importing config as profile '$PROFILE_NAME'..."
  lango config import "$CONFIG_SECRET" --profile "$PROFILE_NAME"
  echo "[$AGENT_NAME] Config imported."
fi

# ── Inject wallet private key as encrypted secret ──
# Re-create keyfile because bootstrap shreds it after crypto init (config import).
if [ -n "$AGENT_PRIVATE_KEY" ]; then
  if [ -f "$PASSPHRASE_SECRET" ]; then
    cp "$PASSPHRASE_SECRET" "$LANGO_DIR/keyfile"
    chmod 600 "$LANGO_DIR/keyfile"
  fi
  echo "[$AGENT_NAME] Storing wallet private key..."
  lango security secrets set wallet.privatekey --value-hex "$AGENT_PRIVATE_KEY"
  echo "[$AGENT_NAME] Wallet key stored."
fi

# Re-create keyfile for `lango serve` bootstrap (shredded by previous commands).
if [ -f "$PASSPHRASE_SECRET" ]; then
  cp "$PASSPHRASE_SECRET" "$LANGO_DIR/keyfile"
  chmod 600 "$LANGO_DIR/keyfile"
fi

echo "[$AGENT_NAME] Starting lango..."
exec lango "$@"
