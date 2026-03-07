package contract

import (
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Minimal ERC-20 ABI for testing.
const testERC20ABI = `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]`

func TestABICache_GetSet(t *testing.T) {
	cache := NewABICache()
	addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	// Get on empty cache returns false.
	_, ok := cache.Get(1, addr)
	assert.False(t, ok)

	// Parse and set.
	parsed, err := ParseABI(testERC20ABI)
	require.NoError(t, err)
	cache.Set(1, addr, parsed)

	// Get returns the cached value.
	got, ok := cache.Get(1, addr)
	require.True(t, ok)
	assert.Equal(t, parsed, got)

	// Different chain ID returns false.
	_, ok = cache.Get(2, addr)
	assert.False(t, ok)
}

func TestABICache_GetOrParse(t *testing.T) {
	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: testERC20ABI, wantErr: false},
		{give: "not json", wantErr: true},
		{give: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cache := NewABICache()
			addr := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

			parsed, err := cache.GetOrParse(1, addr, tt.give)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, parsed)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, parsed)

				// Second call should return cached result.
				cached, ok := cache.Get(1, addr)
				require.True(t, ok)
				assert.Equal(t, parsed, cached)
			}
		})
	}
}

func TestABICache_ConcurrentAccess(t *testing.T) {
	cache := NewABICache()
	addr := common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	parsed, err := ParseABI(testERC20ABI)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(chainID int64) {
			defer wg.Done()
			cache.Set(chainID, addr, parsed)
		}(int64(i % 5))
		go func(chainID int64) {
			defer wg.Done()
			cache.Get(chainID, addr)
		}(int64(i % 5))
	}
	wg.Wait()
}
