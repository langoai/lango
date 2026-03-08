#!/bin/bash
set -euo pipefail

# ABI consistency check: Solidity <-> Go bindings
# Usage: ./scripts/check-abi.sh [contracts_dir]

CONTRACTS_DIR="${1:-contracts}"
BINDINGS_DIR="internal/smartaccount/bindings"
ERRORS=0

check_abi() {
    local contract="$1"
    local go_file="$2"
    local go_var="$3"

    echo "Checking $contract <-> $go_file ($go_var)..."

    # Extract Solidity ABI via forge inspect.
    local sol_abi
    sol_abi=$(cd "$CONTRACTS_DIR" && forge inspect "$contract" abi 2>/dev/null) || {
        echo "  WARNING: Cannot inspect $contract (forge not available or contract not found)"
        return
    }

    # Extract Go ABI (the JSON between backticks after the const declaration).
    local go_abi
    go_abi=$(sed -n "/^const ${go_var}/,/^\`$/p" "$BINDINGS_DIR/$go_file" | sed '1d;$d' | tr -d '\t\n ')

    if [ -z "$go_abi" ]; then
        echo "  ERROR: Could not extract $go_var from $BINDINGS_DIR/$go_file"
        ERRORS=$((ERRORS + 1))
        return
    fi

    # Extract method names from both ABIs.
    local sol_methods go_methods
    sol_methods=$(echo "$sol_abi" | jq -r '.[].name // empty' 2>/dev/null | sort)
    go_methods=$(echo "$go_abi" | jq -r '.[].name // empty' 2>/dev/null | sort)

    # Compare method lists.
    local diff_result
    diff_result=$(diff <(echo "$sol_methods") <(echo "$go_methods") || true)

    if [ -z "$diff_result" ]; then
        echo "  OK: Methods match"
    else
        echo "  FAIL: Method mismatch:"
        echo "$diff_result" | sed 's/^/    /'
        ERRORS=$((ERRORS + 1))
    fi
}

# Check each contract binding pair.
check_abi "Safe7579" "safe7579.go" "Safe7579ABI"
check_abi "LangoSessionValidator" "session_validator.go" "SessionValidatorABI"
check_abi "LangoSpendingHook" "spending_hook.go" "SpendingHookABI"
check_abi "LangoEscrowExecutor" "escrow_executor.go" "EscrowExecutorABI"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: $ERRORS ABI mismatch(es) found"
    exit 1
fi

echo ""
echo "OK: All ABI bindings match Solidity contracts"
