package module

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// installModuleSelector is bytes4(keccak256("installModule(uint256,address,bytes)")).
var installModuleSelector = crypto.Keccak256(
	[]byte("installModule(uint256,address,bytes)"),
)[:4]

// uninstallModuleSelector is bytes4(keccak256("uninstallModule(uint256,address,bytes)")).
var uninstallModuleSelector = crypto.Keccak256(
	[]byte("uninstallModule(uint256,address,bytes)"),
)[:4]

// moduleABIArgs defines the ABI argument types for module install/uninstall.
var moduleABIArgs = abi.Arguments{
	{Type: mustType("uint256")},
	{Type: mustType("address")},
	{Type: mustType("bytes")},
}

// EncodeInstallModule encodes the ERC-7579 installModule call.
//
//	installModule(uint256 moduleType, address module, bytes initData)
func EncodeInstallModule(
	moduleType uint8, moduleAddr common.Address, initData []byte,
) ([]byte, error) {
	packed, err := moduleABIArgs.Pack(
		new(big.Int).SetUint64(uint64(moduleType)),
		moduleAddr,
		initData,
	)
	if err != nil {
		return nil, fmt.Errorf("encode installModule: %w", err)
	}
	return append(installModuleSelector, packed...), nil
}

// EncodeUninstallModule encodes the ERC-7579 uninstallModule call.
//
//	uninstallModule(uint256 moduleType, address module, bytes deInitData)
func EncodeUninstallModule(
	moduleType uint8, moduleAddr common.Address, deInitData []byte,
) ([]byte, error) {
	packed, err := moduleABIArgs.Pack(
		new(big.Int).SetUint64(uint64(moduleType)),
		moduleAddr,
		deInitData,
	)
	if err != nil {
		return nil, fmt.Errorf("encode uninstallModule: %w", err)
	}
	return append(uninstallModuleSelector, packed...), nil
}

// mustType creates an ABI type or panics (safe for package init).
func mustType(t string) abi.Type {
	typ, err := abi.NewType(t, "", nil)
	if err != nil {
		panic(fmt.Sprintf("invalid ABI type %q: %v", t, err))
	}
	return typ
}
