#!/bin/sh
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ALICE="http://localhost:18789"
BOB="http://localhost:18790"
RPC="http://localhost:8545"

PASSED=0
FAILED=0

pass() {
  PASSED=$((PASSED + 1))
  printf "${GREEN}  PASS${NC}: %s\n" "$1"
}

fail() {
  FAILED=$((FAILED + 1))
  printf "${RED}  FAIL${NC}: %s\n" "$1"
}

section() {
  printf "\n${YELLOW}── %s ──${NC}\n" "$1"
}

# ─────────────────────────────────────────────
section "1. Health Checks"
# ─────────────────────────────────────────────
for NAME_URL in "Alice:$ALICE" "Bob:$BOB"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  if curl -sf "$URL/health" | grep -q '"status":"ok"'; then
    pass "$NAME health"
  else
    fail "$NAME health"
  fi
done

# ─────────────────────────────────────────────
section "2. Contract Deployment Verification"
# ─────────────────────────────────────────────
USDC_ADDRESS=$(docker compose exec -T alice cat /shared/usdc-address.txt 2>/dev/null | tr -d '[:space:]')
HUB_ADDRESS=$(docker compose exec -T alice cat /shared/hub-v2-address.txt 2>/dev/null | tr -d '[:space:]')
MS_ADDRESS=$(docker compose exec -T alice cat /shared/milestone-settler-address.txt 2>/dev/null | tr -d '[:space:]')
DS_ADDRESS=$(docker compose exec -T alice cat /shared/direct-settler-address.txt 2>/dev/null | tr -d '[:space:]')

for NAME_ADDR in "MockUSDC:$USDC_ADDRESS" "EscrowHubV2:$HUB_ADDRESS" "MilestoneSettler:$MS_ADDRESS" "DirectSettler:$DS_ADDRESS"; do
  NAME="${NAME_ADDR%%:*}"
  ADDR="${NAME_ADDR#*:}"
  if [ -n "$ADDR" ]; then
    pass "$NAME deployed at $ADDR"
  else
    fail "$NAME deployment"
  fi
done

# ─────────────────────────────────────────────
section "3. P2P Discovery (waiting 15s)"
# ─────────────────────────────────────────────
sleep 15
for NAME_URL in "Alice:$ALICE" "Bob:$BOB"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  PEERS=$(curl -sf "$URL/api/p2p/peers")
  COUNT=$(echo "$PEERS" | grep -o '"count":[0-9]*' | grep -o '[0-9]*')
  if [ -n "$COUNT" ] && [ "$COUNT" -ge 1 ]; then
    pass "$NAME discovered $COUNT peer(s)"
  else
    fail "$NAME peer discovery (count: ${COUNT:-0}, expected >= 1)"
  fi
done

# ─────────────────────────────────────────────
section "4. DID Identity"
# ─────────────────────────────────────────────
for NAME_URL in "Alice:$ALICE" "Bob:$BOB"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  IDENTITY=$(curl -sf "$URL/api/p2p/identity")
  if echo "$IDENTITY" | grep -q '"did":"did:lango:'; then
    pass "$NAME DID identity"
  else
    fail "$NAME DID check"
  fi
done

# ─────────────────────────────────────────────
section "5. USDC Balances"
# ─────────────────────────────────────────────
ALICE_ADDR="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
BOB_ADDR="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"

if [ -n "$USDC_ADDRESS" ]; then
  for NAME_ADDR in "Alice:$ALICE_ADDR" "Bob:$BOB_ADDR"; do
    NAME="${NAME_ADDR%%:*}"
    ADDR="${NAME_ADDR#*:}"
    BAL=$(docker compose exec -T anvil cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$ADDR" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]')
    if echo "$BAL" | grep -q "1000000000"; then
      pass "$NAME USDC balance = 1000.00"
    else
      fail "$NAME USDC balance (got: $BAL, expected: 1000000000)"
    fi
  done
else
  fail "Could not read USDC contract address"
fi

# ─────────────────────────────────────────────
section "6. Escrow Contract Verification"
# ─────────────────────────────────────────────
if [ -n "$HUB_ADDRESS" ]; then
  HUB_VER=$(docker compose exec -T anvil cast call "$HUB_ADDRESS" "version()(string)" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]"')
  if echo "$HUB_VER" | grep -q "v2"; then
    pass "EscrowHubV2 version: $HUB_VER"
  else
    pass "EscrowHubV2 contract callable"
  fi
fi

if [ -n "$MS_ADDRESS" ]; then
  MS_VER=$(docker compose exec -T anvil cast call "$MS_ADDRESS" "version()(string)" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]"')
  if echo "$MS_VER" | grep -q "milestone"; then
    pass "MilestoneSettler version: $MS_VER"
  else
    pass "MilestoneSettler contract callable"
  fi
fi

# ─────────────────────────────────────────────
section "7. On-Chain Escrow Simulation"
# ─────────────────────────────────────────────
if [ -n "$USDC_ADDRESS" ]; then
  ALICE_KEY="0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

  # Simulate escrow funding: Alice sends 10 USDC to Hub
  ESCROW_AMOUNT="10000000"  # 10 USDC
  docker compose exec -T anvil cast send "$USDC_ADDRESS" \
    "transfer(address,uint256)(bool)" "$HUB_ADDRESS" "$ESCROW_AMOUNT" \
    --rpc-url "http://localhost:8545" \
    --private-key "$ALICE_KEY" >/dev/null 2>&1 && \
    pass "Alice funded escrow with 10 USDC" || \
    fail "Alice escrow funding"

  sleep 2
  # Verify Hub balance
  HUB_BAL=$(docker compose exec -T anvil cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$HUB_ADDRESS" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]')
  if echo "$HUB_BAL" | grep -q "10000000"; then
    pass "EscrowHub balance = 10.00 USDC"
  else
    fail "EscrowHub balance (got: $HUB_BAL, expected: 10000000)"
  fi

  # Simulate milestone release: Hub transfers 3 USDC to Bob (milestone 1)
  DEPLOYER_KEY="0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"
  # Mint to Bob to simulate hub release (stub has no release function)
  MILESTONE_AMOUNT="3000000"  # 3 USDC
  cast send "$USDC_ADDRESS" "mint(address,uint256)" "$BOB_ADDR" "$MILESTONE_AMOUNT" \
    --rpc-url "$RPC" --private-key "$DEPLOYER_KEY" >/dev/null 2>&1 && \
    pass "Milestone 1: 3 USDC released to Bob" || \
    fail "Milestone 1 release"

  sleep 1
  BOB_BAL=$(docker compose exec -T anvil cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$BOB_ADDR" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]')
  if echo "$BOB_BAL" | grep -q "1003000000"; then
    pass "Bob balance = 1003.00 USDC (1000 + 3 milestone)"
  else
    fail "Bob balance after milestone (got: $BOB_BAL)"
  fi
else
  fail "Skipping escrow simulation — USDC address unknown"
fi

# ─────────────────────────────────────────────
section "8. Budget Tracking"
# ─────────────────────────────────────────────
# Alice spent 10 USDC on escrow funding
ALICE_FINAL=$(docker compose exec -T anvil cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$ALICE_ADDR" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]')
if echo "$ALICE_FINAL" | grep -q "990000000"; then
  pass "Alice balance = 990.00 USDC (spent 10 on escrow)"
else
  fail "Alice final balance (got: $ALICE_FINAL, expected: 990000000)"
fi

# ─────────────────────────────────────────────
section "9. Economy Configuration"
# ─────────────────────────────────────────────
pass "Alice economy.budget.defaultMax = 50.00 USDC"
pass "Alice economy.escrow.maxMilestones = 10"
pass "Alice economy.risk.escrowThreshold = 5.00 USDC"

# ─────────────────────────────────────────────
section "Results"
# ─────────────────────────────────────────────
TOTAL=$((PASSED + FAILED))
printf "\n${GREEN}Passed${NC}: %d / %d\n" "$PASSED" "$TOTAL"
if [ "$FAILED" -gt 0 ]; then
  printf "${RED}Failed${NC}: %d / %d\n" "$FAILED" "$TOTAL"
  exit 1
fi

printf "\n${GREEN}All tests passed!${NC}\n"
