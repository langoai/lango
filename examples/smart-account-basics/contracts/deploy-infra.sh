#!/bin/sh
set -e

DEPLOYER_KEY="0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"
RPC="http://anvil:8545"
AGENT_ADDR="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

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
  --rpc-url "$RPC" \
  --private-key "$DEPLOYER_KEY" \
  --broadcast 2>&1)
echo "$DEPLOY_OUTPUT"
USDC_ADDRESS=$(echo "$DEPLOY_OUTPUT" | grep -i "deployed to" | grep -o '0x[0-9a-fA-F]\{40\}')
echo -n "$USDC_ADDRESS" > /shared/usdc-address.txt
echo "[setup] MockUSDC at: $USDC_ADDRESS"

# Mint 1000 USDC to agent
AMOUNT="1000000000"
echo "[setup] Minting 1000 USDC to agent..."
cast send "$USDC_ADDRESS" "mint(address,uint256)" "$AGENT_ADDR" "$AMOUNT" \
  --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" >/dev/null

# Deploy EntryPoint stub
echo "[setup] Deploying EntryPoint stub..."
EP_OUTPUT=$(forge create /contracts/EntryPointStub.sol:EntryPointStub \
  --rpc-url "$RPC" \
  --private-key "$DEPLOYER_KEY" \
  --broadcast 2>&1)
echo "$EP_OUTPUT"
EP_ADDRESS=$(echo "$EP_OUTPUT" | grep -i "deployed to" | grep -o '0x[0-9a-fA-F]\{40\}')
echo -n "$EP_ADDRESS" > /shared/entrypoint-address.txt
echo "[setup] EntryPoint at: $EP_ADDRESS"

# Deploy Factory stub
echo "[setup] Deploying Factory stub..."
FACTORY_OUTPUT=$(forge create /contracts/FactoryStub.sol:FactoryStub \
  --rpc-url "$RPC" \
  --private-key "$DEPLOYER_KEY" \
  --broadcast 2>&1)
echo "$FACTORY_OUTPUT"
FACTORY_ADDRESS=$(echo "$FACTORY_OUTPUT" | grep -i "deployed to" | grep -o '0x[0-9a-fA-F]\{40\}')
echo -n "$FACTORY_ADDRESS" > /shared/factory-address.txt
echo "[setup] Factory at: $FACTORY_ADDRESS"

echo "[setup] Done."
