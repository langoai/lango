package contract

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ABICache is a thread-safe cache for parsed contract ABIs.
type ABICache struct {
	mu    sync.RWMutex
	cache map[string]*abi.ABI
}

// NewABICache creates a new ABI cache.
func NewABICache() *ABICache {
	return &ABICache{
		cache: make(map[string]*abi.ABI),
	}
}

// Get retrieves a cached ABI for the given chain and address.
func (c *ABICache) Get(chainID int64, address common.Address) (*abi.ABI, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	parsed, ok := c.cache[cacheKey(chainID, address)]
	return parsed, ok
}

// Set stores a parsed ABI in the cache.
func (c *ABICache) Set(chainID int64, address common.Address, parsed *abi.ABI) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[cacheKey(chainID, address)] = parsed
}

// GetOrParse retrieves a cached ABI or parses the JSON and caches the result.
func (c *ABICache) GetOrParse(chainID int64, address common.Address, abiJSON string) (*abi.ABI, error) {
	if parsed, ok := c.Get(chainID, address); ok {
		return parsed, nil
	}

	parsed, err := ParseABI(abiJSON)
	if err != nil {
		return nil, fmt.Errorf("parse ABI for %s: %w", address.Hex(), err)
	}

	c.Set(chainID, address, parsed)
	return parsed, nil
}

func cacheKey(chainID int64, address common.Address) string {
	return fmt.Sprintf("%d:%s", chainID, address.Hex())
}
