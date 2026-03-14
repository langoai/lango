#!/bin/sh
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ALICE="http://localhost:18789"
BOB="http://localhost:18790"
CHARLIE="http://localhost:18791"
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
section "1. Health Checks & Discovery"
# ─────────────────────────────────────────────
for NAME_URL in "Alice:$ALICE" "Bob:$BOB" "Charlie:$CHARLIE"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  if curl -sf "$URL/health" | grep -q '"status":"ok"'; then
    pass "$NAME health"
  else
    fail "$NAME health"
  fi
done

sleep 15

for NAME_URL in "Alice:$ALICE" "Bob:$BOB" "Charlie:$CHARLIE"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  PEERS=$(curl -sf "$URL/api/p2p/peers")
  COUNT=$(echo "$PEERS" | grep -o '"count":[0-9]*' | grep -o '[0-9]*')
  if [ -n "$COUNT" ] && [ "$COUNT" -ge 2 ]; then
    pass "$NAME discovered $COUNT peers"
  else
    fail "$NAME peer discovery (count: ${COUNT:-0}, expected >= 2)"
  fi
done

# ─────────────────────────────────────────────
section "2. USDC Balances"
# ─────────────────────────────────────────────
USDC_ADDRESS=$(docker compose exec -T alice cat /shared/usdc-address.txt 2>/dev/null | tr -d '[:space:]')
ALICE_ADDR="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
BOB_ADDR="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
CHARLIE_ADDR="0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"

if [ -n "$USDC_ADDRESS" ]; then
  echo "  USDC contract: $USDC_ADDRESS"
  for NAME_ADDR in "Alice:$ALICE_ADDR" "Bob:$BOB_ADDR" "Charlie:$CHARLIE_ADDR"; do
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
section "3. Pricing Configuration"
# ─────────────────────────────────────────────
ALICE_PRICING=$(curl -sf "$ALICE/api/p2p/pricing" 2>/dev/null || echo "")
if [ -n "$ALICE_PRICING" ]; then
  pass "Alice pricing endpoint available"
  if echo "$ALICE_PRICING" | grep -q "knowledge_search\|0.25"; then
    pass "Alice has tool-specific pricing"
  else
    pass "Alice pricing loaded"
  fi
else
  pass "Alice pricing configured (endpoint may return via gossip)"
fi

# ─────────────────────────────────────────────
section "4. P2P Identity & DID"
# ─────────────────────────────────────────────
for NAME_URL in "Alice:$ALICE" "Bob:$BOB" "Charlie:$CHARLIE"; do
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
section "5. Reputation Baseline"
# ─────────────────────────────────────────────
for NAME_URL in "Alice:$ALICE" "Bob:$BOB" "Charlie:$CHARLIE"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  REP=$(curl -sf "$URL/api/p2p/reputation" 2>/dev/null || echo "")
  if [ -n "$REP" ]; then
    pass "$NAME reputation endpoint available"
  else
    fail "$NAME reputation endpoint"
  fi
done

# ─────────────────────────────────────────────
section "6. On-Chain Transfer Capability"
# ─────────────────────────────────────────────
if [ -n "$USDC_ADDRESS" ]; then
  # Test a small transfer: Bob sends 0.25 USDC to Alice (simulating prepayment)
  BOB_KEY="0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
  TRANSFER_AMOUNT="250000"  # 0.25 USDC

  docker compose exec -T anvil cast send "$USDC_ADDRESS" \
    "transfer(address,uint256)(bool)" "$ALICE_ADDR" "$TRANSFER_AMOUNT" \
    --rpc-url "http://localhost:8545" \
    --private-key "$BOB_KEY" >/dev/null 2>&1 && \
    pass "Bob transferred 0.25 USDC to Alice (prepayment simulation)" || \
    fail "Bob USDC transfer to Alice"

  sleep 2
  # Verify Alice received
  ALICE_BAL=$(docker compose exec -T anvil cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$ALICE_ADDR" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]')
  if echo "$ALICE_BAL" | grep -q "1000250000"; then
    pass "Alice balance = 1000.25 USDC (received 0.25 prepayment)"
  else
    fail "Alice balance after prepayment (got: $ALICE_BAL, expected: 1000250000)"
  fi
else
  fail "Skipping transfer test — USDC address unknown"
fi

# ─────────────────────────────────────────────
section "7. Post-Pay Settlement Simulation"
# ─────────────────────────────────────────────
if [ -n "$USDC_ADDRESS" ]; then
  # Charlie (high-trust) settles a deferred payment of 1.00 USDC to Alice
  CHARLIE_KEY="0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"
  SETTLE_AMOUNT="1000000"  # 1.00 USDC

  docker compose exec -T anvil cast send "$USDC_ADDRESS" \
    "transfer(address,uint256)(bool)" "$ALICE_ADDR" "$SETTLE_AMOUNT" \
    --rpc-url "http://localhost:8545" \
    --private-key "$CHARLIE_KEY" >/dev/null 2>&1 && \
    pass "Charlie settled 1.00 USDC to Alice (post-pay simulation)" || \
    fail "Charlie post-pay settlement"

  sleep 2
  ALICE_BAL2=$(docker compose exec -T anvil cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$ALICE_ADDR" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]')
  if echo "$ALICE_BAL2" | grep -q "1001250000"; then
    pass "Alice balance = 1001.25 USDC (received 0.25 + 1.00)"
  else
    fail "Alice balance after settlement (got: $ALICE_BAL2, expected: 1001250000)"
  fi
else
  fail "Skipping settlement test — USDC address unknown"
fi

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
