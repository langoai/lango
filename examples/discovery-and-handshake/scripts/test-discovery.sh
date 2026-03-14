#!/bin/sh
set -e

# Colors for test output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ALICE="http://localhost:18789"
BOB="http://localhost:18790"

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
section "2. P2P Status"
# ─────────────────────────────────────────────
for NAME_URL in "Alice:$ALICE" "Bob:$BOB"; do
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
section "3. mDNS Discovery (waiting 15s)"
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
    pass "$NAME DID starts with did:lango:"
  else
    fail "$NAME DID check ($IDENTITY)"
  fi
done

# ─────────────────────────────────────────────
section "5. Gossip Agent Card Discovery"
# ─────────────────────────────────────────────
# Alice should see Bob's capabilities and vice versa
ALICE_PEERS=$(curl -sf "$ALICE/api/p2p/peers")
if echo "$ALICE_PEERS" | grep -q '"peers"'; then
  pass "Alice sees peer list via gossip"
else
  fail "Alice gossip peer list"
fi

BOB_PEERS=$(curl -sf "$BOB/api/p2p/peers")
if echo "$BOB_PEERS" | grep -q '"peers"'; then
  pass "Bob sees peer list via gossip"
else
  fail "Bob gossip peer list"
fi

# ─────────────────────────────────────────────
section "6. Handshake Session"
# ─────────────────────────────────────────────
# Check that peers have session information after handshake
ALICE_PEERS_DETAIL=$(curl -sf "$ALICE/api/p2p/peers")
if echo "$ALICE_PEERS_DETAIL" | grep -q '"did"'; then
  pass "Alice has peer DID info (handshake completed)"
else
  fail "Alice peer DID info (handshake may not have completed)"
fi

BOB_PEERS_DETAIL=$(curl -sf "$BOB/api/p2p/peers")
if echo "$BOB_PEERS_DETAIL" | grep -q '"did"'; then
  pass "Bob has peer DID info (handshake completed)"
else
  fail "Bob peer DID info (handshake may not have completed)"
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
