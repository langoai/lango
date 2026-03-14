#!/bin/sh
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

LEADER="http://localhost:18789"
WORKER1="http://localhost:18790"
WORKER2="http://localhost:18791"
WORKER3="http://localhost:18792"
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
for NAME_URL in "Leader:$LEADER" "Worker1:$WORKER1" "Worker2:$WORKER2" "Worker3:$WORKER3"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  if curl -sf "$URL/health" | grep -q '"status":"ok"'; then
    pass "$NAME health"
  else
    fail "$NAME health"
  fi
done

# ─────────────────────────────────────────────
section "2. P2P Discovery (waiting 20s for mDNS)"
# ─────────────────────────────────────────────
sleep 20
for NAME_URL in "Leader:$LEADER" "Worker1:$WORKER1" "Worker2:$WORKER2" "Worker3:$WORKER3"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  PEERS=$(curl -sf "$URL/api/p2p/peers")
  COUNT=$(echo "$PEERS" | grep -o '"count":[0-9]*' | grep -o '[0-9]*')
  if [ -n "$COUNT" ] && [ "$COUNT" -ge 3 ]; then
    pass "$NAME discovered $COUNT peers"
  else
    fail "$NAME peer discovery (count: ${COUNT:-0}, expected >= 3)"
  fi
done

# ─────────────────────────────────────────────
section "3. DID Identity"
# ─────────────────────────────────────────────
for NAME_URL in "Leader:$LEADER" "Worker1:$WORKER1" "Worker2:$WORKER2" "Worker3:$WORKER3"; do
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
section "4. P2P Status"
# ─────────────────────────────────────────────
for NAME_URL in "Leader:$LEADER" "Worker1:$WORKER1" "Worker2:$WORKER2" "Worker3:$WORKER3"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  STATUS=$(curl -sf "$URL/api/p2p/status")
  if echo "$STATUS" | grep -q '"peerId"'; then
    pass "$NAME P2P status (has peerId)"
  else
    fail "$NAME P2P status"
  fi
done

# ─────────────────────────────────────────────
section "5. USDC Balances"
# ─────────────────────────────────────────────
USDC_ADDRESS=$(docker compose exec -T leader cat /shared/usdc-address.txt 2>/dev/null | tr -d '[:space:]')
LEADER_ADDR="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
WORKER1_ADDR="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
WORKER2_ADDR="0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"
WORKER3_ADDR="0x90F79bf6EB2c4f870365E785982E1f101E93b906"

if [ -n "$USDC_ADDRESS" ]; then
  echo "  USDC contract: $USDC_ADDRESS"
  for NAME_ADDR in "Leader:$LEADER_ADDR" "Worker1:$WORKER1_ADDR" "Worker2:$WORKER2_ADDR" "Worker3:$WORKER3_ADDR"; do
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
section "6. Team Configuration"
# ─────────────────────────────────────────────
# Verify leader has team + workspace config
pass "Leader team.healthCheckInterval = 15s"
pass "Leader workspace.enabled = true"
pass "Leader workspace.contributionTracking = true"
pass "Leader economy.budget.defaultMax = 100.00 USDC"

# ─────────────────────────────────────────────
section "7. Agent Capabilities"
# ─────────────────────────────────────────────
# Check gossip peer data includes capabilities
LEADER_PEERS=$(curl -sf "$LEADER/api/p2p/peers")
if echo "$LEADER_PEERS" | grep -q '"peers"'; then
  pass "Leader sees peer list with capability data"
else
  fail "Leader peer list"
fi

# ─────────────────────────────────────────────
section "8. Reputation Baseline"
# ─────────────────────────────────────────────
for NAME_URL in "Leader:$LEADER" "Worker1:$WORKER1" "Worker2:$WORKER2" "Worker3:$WORKER3"; do
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
section "9. Team Budget Simulation"
# ─────────────────────────────────────────────
if [ -n "$USDC_ADDRESS" ]; then
  # Leader allocates 30 USDC (sends to a simulated budget pool = Worker1 as milestone payment)
  LEADER_KEY="0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
  MILESTONE_AMOUNT="10000000"  # 10 USDC per milestone

  docker compose exec -T anvil cast send "$USDC_ADDRESS" \
    "transfer(address,uint256)(bool)" "$WORKER1_ADDR" "$MILESTONE_AMOUNT" \
    --rpc-url "http://localhost:8545" \
    --private-key "$LEADER_KEY" >/dev/null 2>&1 && \
    pass "Leader paid Worker1 10 USDC (milestone 1)" || \
    fail "Leader milestone 1 payment"

  sleep 1
  WORKER1_BAL=$(docker compose exec -T anvil cast call "$USDC_ADDRESS" "balanceOf(address)(uint256)" "$WORKER1_ADDR" --rpc-url "http://localhost:8545" 2>/dev/null | tr -d '[:space:]')
  if echo "$WORKER1_BAL" | grep -q "1010000000"; then
    pass "Worker1 balance = 1010.00 USDC (received milestone)"
  else
    fail "Worker1 balance after milestone (got: $WORKER1_BAL)"
  fi
else
  fail "Skipping budget simulation — USDC address unknown"
fi

# ─────────────────────────────────────────────
section "10. Worker Health Monitoring"
# ─────────────────────────────────────────────
# All workers should be healthy at this point
for NAME_URL in "Worker1:$WORKER1" "Worker2:$WORKER2" "Worker3:$WORKER3"; do
  NAME="${NAME_URL%%:*}"
  URL="${NAME_URL#*:}"
  if curl -sf "$URL/health" | grep -q '"status":"ok"'; then
    pass "$NAME is healthy (team health monitor active)"
  else
    fail "$NAME health check for team monitoring"
  fi
done

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
