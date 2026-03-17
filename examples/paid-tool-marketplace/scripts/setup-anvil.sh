#!/bin/sh
set -e

# Anvil deterministic addresses (accounts[0..2])
ALICE_ADDR="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
BOB_ADDR="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
CHARLIE_ADDR="0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"

# Deployer = account[9] (last Anvil account — not used by agents)
DEPLOYER_KEY="0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"

RPC="http://anvil:8545"

# Suppress nightly warnings from Foundry.
export FOUNDRY_DISABLE_NIGHTLY_WARNING=1

# Use writable directories for forge compilation output and cache.
export FOUNDRY_OUT="/tmp/forge-out"
export FOUNDRY_CACHE_PATH="/tmp/forge-cache"
mkdir -p "$FOUNDRY_OUT" "$FOUNDRY_CACHE_PATH"

echo "[setup] Waiting for Anvil..."
until cast block-number --rpc-url "$RPC" >/dev/null 2>&1; do sleep 1; done
echo "[setup] Anvil is ready."

# Deploy MockUSDC (non-JSON output is more reliable for parsing)
echo "[setup] Deploying MockUSDC..."
DEPLOY_OUTPUT=$(forge create /contracts/MockUSDC.sol:MockUSDC \
  --rpc-url "$RPC" \
  --private-key "$DEPLOYER_KEY" \
  --broadcast 2>&1)

echo "[setup] Deploy output:"
echo "$DEPLOY_OUTPUT"

# Extract "Deployed to: 0x..." from forge's human-readable output.
USDC_ADDRESS=$(echo "$DEPLOY_OUTPUT" | grep -i "deployed to" | grep -o '0x[0-9a-fA-F]\{40\}')

if [ -z "$USDC_ADDRESS" ]; then
  echo "[setup] ERROR: Failed to extract USDC address"
  exit 1
fi

echo "[setup] MockUSDC deployed at: $USDC_ADDRESS"
echo -n "$USDC_ADDRESS" > /shared/usdc-address.txt

# Mint 1000 USDC (1000 * 10^6 = 1000000000) to each agent
AMOUNT="1000000000"

echo "[setup] Minting 1000 USDC to Alice..."
cast send "$USDC_ADDRESS" "mint(address,uint256)" "$ALICE_ADDR" "$AMOUNT" \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" >/dev/null

echo "[setup] Minting 1000 USDC to Bob..."
cast send "$USDC_ADDRESS" "mint(address,uint256)" "$BOB_ADDR" "$AMOUNT" \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" >/dev/null

echo "[setup] Minting 1000 USDC to Charlie..."
cast send "$USDC_ADDRESS" "mint(address,uint256)" "$CHARLIE_ADDR" "$AMOUNT" \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" >/dev/null

# Verify balances
for ADDR in "$ALICE_ADDR" "$BOB_ADDR" "$CHARLIE_ADDR"; do
  BAL=$(cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$ADDR" --rpc-url "$RPC")
  echo "[setup] Balance of $ADDR: $BAL"
done

echo "[setup] Done."
