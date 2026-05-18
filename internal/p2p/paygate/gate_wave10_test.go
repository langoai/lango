package paygate

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGateLedgerWave10_ReturnsDeferredLedger(t *testing.T) {
	t.Parallel()

	gate := testGate(func(toolName string) (string, bool) {
		return "0.50", false
	})

	id := gate.Ledger().Add("did:peer:ledger", "paid-tool", "0.50")
	pending := gate.Ledger().PendingByPeer("did:peer:ledger")

	require.Len(t, pending, 1)
	assert.Equal(t, id, pending[0].ID)
	assert.Same(t, gate.ledger, gate.Ledger())
}

func TestParseAuthorizationWave10_AcceptsStringAndJSONNumberFields(t *testing.T) {
	t.Parallel()

	authMap := makeValidAuth("0x1234567890abcdef1234567890abcdef12345678", big.NewInt(500000))
	authMap["value"] = "0x7a120"
	authMap["validAfter"] = float64(12)
	authMap["validBefore"] = float64(34)
	authMap["v"] = float64(28)

	auth, err := parseAuthorization(authMap)
	require.NoError(t, err)

	assert.Equal(t, common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), auth.From)
	assert.Equal(t, common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"), auth.To)
	assert.Equal(t, big.NewInt(500000), auth.Value)
	assert.Equal(t, big.NewInt(12), auth.ValidAfter)
	assert.Equal(t, big.NewInt(34), auth.ValidBefore)
	assert.Equal(t, uint8(28), auth.V)
	assert.Equal(t, byte(1), auth.Nonce[31])
	assert.Equal(t, byte(2), auth.R[31])
	assert.Equal(t, byte(3), auth.S[31])
}

func TestParseAuthorizationWave10_ReportsFieldSpecificErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(map[string]interface{})
		want   string
	}{
		{
			name: "missing from",
			mutate: func(m map[string]interface{}) {
				delete(m, "from")
			},
			want: `from: missing field "from"`,
		},
		{
			name: "invalid to address",
			mutate: func(m map[string]interface{}) {
				m["to"] = "not-an-address"
			},
			want: `to: field "to" is not a valid hex address`,
		},
		{
			name: "invalid value",
			mutate: func(m map[string]interface{}) {
				m["value"] = "not-an-int"
			},
			want: `value: field "value": invalid integer "not-an-int"`,
		},
		{
			name: "short nonce",
			mutate: func(m map[string]interface{}) {
				m["nonce"] = "0x1234"
			},
			want: `nonce: field "nonce": expected 32 bytes`,
		},
		{
			name: "v out of range",
			mutate: func(m map[string]interface{}) {
				m["v"] = float64(256)
			},
			want: `v: field "v" out of uint8 range`,
		},
		{
			name: "r wrong type",
			mutate: func(m map[string]interface{}) {
				m["r"] = 123
			},
			want: `r: field "r" is not a string`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authMap := makeValidAuth("0x1234567890abcdef1234567890abcdef12345678", big.NewInt(500000))
			tt.mutate(authMap)

			auth, err := parseAuthorization(authMap)
			require.Error(t, err)
			assert.Nil(t, auth)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestHelperExtractorsWave10_Edges(t *testing.T) {
	t.Parallel()

	t.Run("getHexAddress rejects non-string", func(t *testing.T) {
		t.Parallel()

		_, err := getHexAddress(map[string]interface{}{"addr": 123}, "addr")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a string")
	})

	t.Run("getBigInt rejects unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := getBigInt(map[string]interface{}{"n": true}, "n")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type bool")
	})

	t.Run("getBytes32 accepts no-prefix hex", func(t *testing.T) {
		t.Parallel()

		got, err := getBytes32(map[string]interface{}{
			"b": "00000000000000000000000000000000000000000000000000000000000000ff",
		}, "b")
		require.NoError(t, err)
		assert.Equal(t, byte(0xff), got[31])
	})

	t.Run("getUint8 rejects missing and non-number fields", func(t *testing.T) {
		t.Parallel()

		_, err := getUint8(map[string]interface{}{}, "v")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing field")

		_, err = getUint8(map[string]interface{}{"v": "27"}, "v")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a number")
	})
}

func TestDeferredLedgerCleanupWave10_RemovesOnlySettledEntries(t *testing.T) {
	t.Parallel()

	ledger := NewDeferredLedger()
	settledA := ledger.Add("did:peer:a", "tool-a", "0.10")
	pending := ledger.Add("did:peer:b", "tool-b", "0.20")
	settledC := ledger.Add("did:peer:c", "tool-c", "0.30")
	require.True(t, ledger.Settle(settledA, "0xaaa"))
	require.True(t, ledger.Settle(settledC, "0xccc"))

	removed := ledger.Cleanup()

	assert.Equal(t, 2, removed)
	gotPending := ledger.Pending()
	require.Len(t, gotPending, 1)
	assert.Equal(t, pending, gotPending[0].ID)
	assert.Equal(t, 0, ledger.Cleanup())
}
