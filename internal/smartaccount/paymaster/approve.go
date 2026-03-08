package paymaster

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ERC-20 approve(address,uint256) selector
var approveSelector = crypto.Keccak256([]byte("approve(address,uint256)"))[:4]

// BuildApproveCalldata builds ERC-20 approve(spender, amount) calldata.
func BuildApproveCalldata(spender common.Address, amount *big.Int) []byte {
	data := make([]byte, 0, 68)
	data = append(data, approveSelector...)

	// spender address (left-padded to 32 bytes)
	spenderPadded := make([]byte, 32)
	copy(spenderPadded[12:], spender.Bytes())
	data = append(data, spenderPadded...)

	// amount (left-padded to 32 bytes)
	amountPadded := make([]byte, 32)
	if amount != nil {
		b := amount.Bytes()
		copy(amountPadded[32-len(b):], b)
	}
	data = append(data, amountPadded...)

	return data
}

// ApprovalCall represents a contract call to approve tokens.
type ApprovalCall struct {
	TokenAddress    common.Address
	PaymasterAddr   common.Address
	Amount          *big.Int
	ApproveCalldata []byte
}

// NewApprovalCall creates an ERC-20 approve call for the paymaster.
func NewApprovalCall(tokenAddr, paymasterAddr common.Address, amount *big.Int) *ApprovalCall {
	return &ApprovalCall{
		TokenAddress:    tokenAddr,
		PaymasterAddr:   paymasterAddr,
		Amount:          amount,
		ApproveCalldata: BuildApproveCalldata(paymasterAddr, amount),
	}
}
