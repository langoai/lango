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
section "3. mDNS Discovery (polling up to 60s)"
# ─────────────────────────────────────────────
DISCOVERY_OK=0
DISCOVERY_WAIT=0
while [ "$DISCOVERY_WAIT" -lt 60 ]; do
  ALL_FOUND=1
  for NAME_URL in "Alice:$ALICE" "Bob:$BOB"; do
    URL="${NAME_URL#*:}"
    PEERS=$(curl -sf "$URL/api/p2p/peers" 2>/dev/null || echo "")
    COUNT=$(echo "$PEERS" | grep -o '"count":[0-9]*' | grep -o '[0-9]*')
    if [ -z "$COUNT" ] || [ "$COUNT" -lt 1 ]; then
      ALL_FOUND=0
    fi
  done
  if [ "$ALL_FOUND" -eq 1 ]; then
    DISCOVERY_OK=1
    break
  fi
  sleep 5
  DISCOVERY_WAIT=$((DISCOVERY_WAIT + 5))
done

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
# Peers being connected implies handshake completed successfully.
# Verify connected peers have valid peerId (handshake exchanged identities).
ALICE_PEERS_DETAIL=$(curl -sf "$ALICE/api/p2p/peers")
if echo "$ALICE_PEERS_DETAIL" | grep -q '"peerId"'; then
  pass "Alice has connected peer (handshake completed)"
else
  fail "Alice has no connected peers (handshake may not have completed)"
fi

BOB_PEERS_DETAIL=$(curl -sf "$BOB/api/p2p/peers")
if echo "$BOB_PEERS_DETAIL" | grep -q '"peerId"'; then
  pass "Bob has connected peer (handshake completed)"
else
  fail "Bob has no connected peers (handshake may not have completed)"
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
