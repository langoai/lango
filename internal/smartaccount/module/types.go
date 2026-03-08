package module

import (
	"github.com/ethereum/go-ethereum/common"

	sa "github.com/langoai/lango/internal/smartaccount"
)

// ModuleDescriptor describes an available ERC-7579 module.
type ModuleDescriptor struct {
	Name     string         `json:"name"`
	Address  common.Address `json:"address"`
	Type     sa.ModuleType  `json:"type"`
	Version  string         `json:"version"`
	InitData []byte         `json:"initData,omitempty"`
}
