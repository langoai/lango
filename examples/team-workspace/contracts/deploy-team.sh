#!/bin/sh
set -e

DEPLOYER_KEY="0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"
RPC="http://anvil:8545"
LEADER_ADDR="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
WORKER1_ADDR="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
WORKER2_ADDR="0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"
WORKER3_ADDR="0x90F79bf6EB2c4f870365E785982E1f101E93b906"

export FOUNDRY_DISABLE_NIGHTLY_WARNING=1
export FOUNDRY_OUT="/tmp/forge-out"
export FOUNDRY_CACHE_PATH="/tmp/forge-cache"
mkdir -p "$FOUNDRY_OUT" "$FOUNDRY_CACHE_PATH"

echo "[setup] Waiting for Anvil..."
until cast block-number --rpc-url "$RPC" >/dev/null 2>&1; do sleep 1; done
echo "[setup] Anvil is ready."

# Deploy MockUSDC
echo "[setup] Deploying MockUSDC..."
DEPLOY_OUTPUT=$(forge create /contracts/MockUSDC.sol:MockUSDC \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" --broadcast 2>&1)
echo "$DEPLOY_OUTPUT"
USDC_ADDRESS=$(echo "$DEPLOY_OUTPUT" | grep -i "deployed to" | grep -o '0x[0-9a-fA-F]\{40\}')
echo -n "$USDC_ADDRESS" > /shared/usdc-address.txt
echo "[setup] MockUSDC at: $USDC_ADDRESS"

# Mint 1000 USDC to each agent
AMOUNT="1000000000"
for ADDR in "$LEADER_ADDR" "$WORKER1_ADDR" "$WORKER2_ADDR" "$WORKER3_ADDR"; do
  echo "[setup] Minting 1000 USDC to $ADDR..."
  cast send "$USDC_ADDRESS" "mint(address,uint256)" "$ADDR" "$AMOUNT" \
    --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" >/dev/null
done

# Verify balances
for ADDR in "$LEADER_ADDR" "$WORKER1_ADDR" "$WORKER2_ADDR" "$WORKER3_ADDR"; do
  BAL=$(cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$ADDR" --rpc-url "$RPC")
  echo "[setup] Balance of $ADDR: $BAL"
done

echo "[setup] Done."
