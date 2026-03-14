#!/bin/sh
set -e

LANGO_DIR="$HOME/.lango"
mkdir -p "$LANGO_DIR"

# ── Wait for setup sidecar to write the USDC contract address ──
echo "[$AGENT_NAME] Waiting for USDC contract address..."
TIMEOUT=60
ELAPSED=0
while [ ! -f /shared/usdc-address.txt ]; do
  sleep 1
  ELAPSED=$((ELAPSED + 1))
  if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
    echo "[$AGENT_NAME] ERROR: Timed out waiting for /shared/usdc-address.txt"
    exit 1
  fi
done
USDC_ADDRESS=$(cat /shared/usdc-address.txt)
echo "[$AGENT_NAME] USDC contract: $USDC_ADDRESS"

# ── Set up passphrase keyfile ──
PASSPHRASE_SECRET="${LANGO_PASSPHRASE_FILE:-/run/secrets/lango_passphrase}"
if [ -f "$PASSPHRASE_SECRET" ]; then
  cp "$PASSPHRASE_SECRET" "$LANGO_DIR/keyfile"
  chmod 600 "$LANGO_DIR/keyfile"
fi

# ── Import config with USDC address substituted ──
CONFIG_SECRET="${LANGO_CONFIG_FILE:-/run/secrets/lango_config}"
PROFILE_NAME="${LANGO_PROFILE:-default}"

if [ -f "$CONFIG_SECRET" ] && [ ! -f "$LANGO_DIR/lango.db" ]; then
  echo "[$AGENT_NAME] Importing config as profile '$PROFILE_NAME'..."
  cp "$CONFIG_SECRET" /tmp/lango-import.json
  # Replace placeholder USDC address with the deployed contract address
  sed -i "s/PLACEHOLDER_USDC_ADDRESS/$USDC_ADDRESS/g" /tmp/lango-import.json
  lango config import /tmp/lango-import.json --profile "$PROFILE_NAME"
  rm -f /tmp/lango-import.json
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
