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

// writeMethod executes a state-changing contract call and returns the transaction hash.
func (c *HubClient) writeMethod(ctx context.Context, method string, args ...interface{}) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  method,
		Args:    args,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", method, err)
	}
	return result.TxHash, nil
}

// readMethod executes a read-only contract call and returns the result data.
func (c *HubClient) readMethod(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
	result, err := c.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  method,
		Args:    args,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", method, err)
	}
	return result.Data, nil
}

// CreateDeal creates a new escrow deal on-chain.
func (c *HubClient) CreateDeal(ctx context.Context, seller, token common.Address, amount, deadline *big.Int) (*big.Int, string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  MethodCreateDeal,
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
	return c.writeMethod(ctx, MethodDeposit, dealID)
}

// SubmitWork submits a work proof hash for a deal.
func (c *HubClient) SubmitWork(ctx context.Context, dealID *big.Int, workHash [32]byte) (string, error) {
	return c.writeMethod(ctx, MethodSubmitWork, dealID, workHash)
}

// Release releases escrow funds to the seller.
func (c *HubClient) Release(ctx context.Context, dealID *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodRelease, dealID)
}

// Refund returns escrow funds to the buyer (after deadline).
func (c *HubClient) Refund(ctx context.Context, dealID *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodRefund, dealID)
}

// Dispute raises a dispute on a deal.
func (c *HubClient) Dispute(ctx context.Context, dealID *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodDispute, dealID)
}

// ResolveDispute resolves a disputed deal via arbitrator.
func (c *HubClient) ResolveDispute(ctx context.Context, dealID *big.Int, sellerFavor bool, sellerAmount, buyerAmount *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodResolveDispute, dealID, sellerFavor, sellerAmount, buyerAmount)
}

// GetDeal reads the on-chain deal state.
func (c *HubClient) GetDeal(ctx context.Context, dealID *big.Int) (*OnChainDeal, error) {
	data, err := c.readMethod(ctx, MethodGetDeal, dealID)
	if err != nil {
		return nil, err
	}
	return parseDealResult(data)
}

// NextDealID reads the next deal ID counter.
func (c *HubClient) NextDealID(ctx context.Context) (*big.Int, error) {
	data, err := c.readMethod(ctx, MethodNextDealID)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if id, ok := data[0].(*big.Int); ok {
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
