package hub

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

// HubClient provides typed access to the LangoEscrowHub contract.
type HubClient struct {
	caller  contract.ContractCaller
	address common.Address
	chainID int64
	abiJSON string
}

// NewHubClient creates a hub client for the given contract address.
func NewHubClient(caller contract.ContractCaller, address common.Address, chainID int64) *HubClient {
	return &HubClient{
		caller:  caller,
		address: address,
		chainID: chainID,
		abiJSON: hubABIJSON,
	}
}

// CreateDeal creates a new escrow deal on-chain.
func (c *HubClient) CreateDeal(ctx context.Context, seller, token common.Address, amount, deadline *big.Int) (*big.Int, string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "createDeal",
		Args:    []interface{}{seller, token, amount, deadline},
	})
	if err != nil {
		return nil, "", fmt.Errorf("create deal: %w", err)
	}

	// Parse dealId from return value (uint256).
	var dealID *big.Int
	if len(result.Data) > 0 {
		if id, ok := result.Data[0].(*big.Int); ok {
			dealID = id
		}
	}
	return dealID, result.TxHash, nil
}

// Deposit deposits ERC-20 tokens into the escrow for a deal.
func (c *HubClient) Deposit(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "deposit",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("deposit deal %s: %w", dealID.String(), err)
	}
	return result.TxHash, nil
}

// SubmitWork submits a work proof hash for a deal.
func (c *HubClient) SubmitWork(ctx context.Context, dealID *big.Int, workHash [32]byte) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "submitWork",
		Args:    []interface{}{dealID, workHash},
	})
	if err != nil {
		return "", fmt.Errorf("submit work deal %s: %w", dealID.String(), err)
	}
	return result.TxHash, nil
}

// Release releases escrow funds to the seller.
func (c *HubClient) Release(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "release",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("release deal %s: %w", dealID.String(), err)
	}
	return result.TxHash, nil
}

// Refund returns escrow funds to the buyer (after deadline).
func (c *HubClient) Refund(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "refund",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("refund deal %s: %w", dealID.String(), err)
	}
	return result.TxHash, nil
}

// Dispute raises a dispute on a deal.
func (c *HubClient) Dispute(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "dispute",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("dispute deal %s: %w", dealID.String(), err)
	}
	return result.TxHash, nil
}

// ResolveDispute resolves a disputed deal via arbitrator.
func (c *HubClient) ResolveDispute(ctx context.Context, dealID *big.Int, sellerFavor bool, sellerAmount, buyerAmount *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "resolveDispute",
		Args:    []interface{}{dealID, sellerFavor, sellerAmount, buyerAmount},
	})
	if err != nil {
		return "", fmt.Errorf("resolve dispute deal %s: %w", dealID.String(), err)
	}
	return result.TxHash, nil
}

// GetDeal reads the on-chain deal state.
func (c *HubClient) GetDeal(ctx context.Context, dealID *big.Int) (*OnChainDeal, error) {
	result, err := c.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "getDeal",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return nil, fmt.Errorf("get deal %s: %w", dealID.String(), err)
	}
	return parseDealResult(result.Data)
}

// NextDealID reads the next deal ID counter.
func (c *HubClient) NextDealID(ctx context.Context) (*big.Int, error) {
	result, err := c.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "nextDealId",
	})
	if err != nil {
		return nil, fmt.Errorf("next deal id: %w", err)
	}
	if len(result.Data) > 0 {
		if id, ok := result.Data[0].(*big.Int); ok {
			return id, nil
		}
	}
	return big.NewInt(0), nil
}

// parseDealResult converts raw ABI output to OnChainDeal.
func parseDealResult(data []interface{}) (*OnChainDeal, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty deal result")
	}

	// Try direct struct assertion.
	if d, ok := data[0].(struct {
		Buyer    common.Address
		Seller   common.Address
		Token    common.Address
		Amount   *big.Int
		Deadline *big.Int
		Status   uint8
		WorkHash [32]byte
	}); ok {
		return &OnChainDeal{
			Buyer:    d.Buyer,
			Seller:   d.Seller,
			Token:    d.Token,
			Amount:   d.Amount,
			Deadline: d.Deadline,
			Status:   OnChainDealStatus(d.Status),
			WorkHash: d.WorkHash,
		}, nil
	}

	return nil, fmt.Errorf("unexpected deal result type: %T", data[0])
}
