#!/bin/sh
set -e

DEPLOYER_KEY="0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"
RPC="http://anvil:8545"
ALICE_ADDR="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
BOB_ADDR="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"

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
echo "[setup] Minting USDC..."
cast send "$USDC_ADDRESS" "mint(address,uint256)" "$ALICE_ADDR" "$AMOUNT" \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" >/dev/null
cast send "$USDC_ADDRESS" "mint(address,uint256)" "$BOB_ADDR" "$AMOUNT" \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" >/dev/null

# Deploy EscrowHubV2 stub
echo "[setup] Deploying EscrowHubV2 stub..."
HUB_OUTPUT=$(forge create /contracts/EscrowHubV2Stub.sol:EscrowHubV2Stub \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" --broadcast 2>&1)
echo "$HUB_OUTPUT"
HUB_ADDRESS=$(echo "$HUB_OUTPUT" | grep -i "deployed to" | grep -o '0x[0-9a-fA-F]\{40\}')
echo -n "$HUB_ADDRESS" > /shared/hub-v2-address.txt
echo "[setup] EscrowHubV2 at: $HUB_ADDRESS"

# Deploy MilestoneSettler stub
echo "[setup] Deploying MilestoneSettler stub..."
MS_OUTPUT=$(forge create /contracts/MilestoneSettlerStub.sol:MilestoneSettlerStub \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" --broadcast 2>&1)
echo "$MS_OUTPUT"
MS_ADDRESS=$(echo "$MS_OUTPUT" | grep -i "deployed to" | grep -o '0x[0-9a-fA-F]\{40\}')
echo -n "$MS_ADDRESS" > /shared/milestone-settler-address.txt
echo "[setup] MilestoneSettler at: $MS_ADDRESS"

# Deploy DirectSettler stub
echo "[setup] Deploying DirectSettler stub..."
DS_OUTPUT=$(forge create /contracts/DirectSettlerStub.sol:DirectSettlerStub \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" --broadcast 2>&1)
echo "$DS_OUTPUT"
DS_ADDRESS=$(echo "$DS_OUTPUT" | grep -i "deployed to" | grep -o '0x[0-9a-fA-F]\{40\}')
echo -n "$DS_ADDRESS" > /shared/direct-settler-address.txt
echo "[setup] DirectSettler at: $DS_ADDRESS"

# Verify balances
for ADDR in "$ALICE_ADDR" "$BOB_ADDR"; do
  BAL=$(cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$ADDR" --rpc-url "$RPC")
  echo "[setup] Balance of $ADDR: $BAL"
done

echo "[setup] Done."
