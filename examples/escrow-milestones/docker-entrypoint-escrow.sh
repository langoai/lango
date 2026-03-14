#!/bin/sh
set -e

LANGO_DIR="$HOME/.lango"
mkdir -p "$LANGO_DIR"

# ── Wait for setup to write contract addresses ──
echo "[$AGENT_NAME] Waiting for contract addresses..."
TIMEOUT=60
ELAPSED=0
while [ ! -f /shared/usdc-address.txt ] || [ ! -f /shared/hub-v2-address.txt ] || \
      [ ! -f /shared/milestone-settler-address.txt ] || [ ! -f /shared/direct-settler-address.txt ]; do
  sleep 1
  ELAPSED=$((ELAPSED + 1))
  if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
    echo "[$AGENT_NAME] ERROR: Timed out waiting for contract addresses"
    exit 1
  fi
done
USDC_ADDRESS=$(cat /shared/usdc-address.txt)
HUB_V2_ADDRESS=$(cat /shared/hub-v2-address.txt)
MS_ADDRESS=$(cat /shared/milestone-settler-address.txt)
DS_ADDRESS=$(cat /shared/direct-settler-address.txt)
echo "[$AGENT_NAME] USDC: $USDC_ADDRESS, Hub: $HUB_V2_ADDRESS"

# ── Set up passphrase keyfile ──
PASSPHRASE_SECRET="${LANGO_PASSPHRASE_FILE:-/run/secrets/lango_passphrase}"
if [ -f "$PASSPHRASE_SECRET" ]; then
  cp "$PASSPHRASE_SECRET" "$LANGO_DIR/keyfile"
  chmod 600 "$LANGO_DIR/keyfile"
fi

# ── Import config with addresses substituted ──
CONFIG_SECRET="${LANGO_CONFIG_FILE:-/run/secrets/lango_config}"
PROFILE_NAME="${LANGO_PROFILE:-default}"

if [ -f "$CONFIG_SECRET" ] && [ ! -f "$LANGO_DIR/lango.db" ]; then
  echo "[$AGENT_NAME] Importing config..."
  cp "$CONFIG_SECRET" /tmp/lango-import.json
  sed -i "s/PLACEHOLDER_USDC_ADDRESS/$USDC_ADDRESS/g" /tmp/lango-import.json
  sed -i "s/PLACEHOLDER_HUB_V2_ADDRESS/$HUB_V2_ADDRESS/g" /tmp/lango-import.json
  sed -i "s/PLACEHOLDER_MILESTONE_SETTLER_ADDRESS/$MS_ADDRESS/g" /tmp/lango-import.json
  sed -i "s/PLACEHOLDER_DIRECT_SETTLER_ADDRESS/$DS_ADDRESS/g" /tmp/lango-import.json
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
