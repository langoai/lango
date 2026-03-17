package hub

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

// HubV2Client provides typed access to the LangoEscrowHubV2 contract.
// It extends V1 methods with V2-specific entry points that support refId.
type HubV2Client struct {
	caller  contract.ContractCaller
	address common.Address
	chainID int64
	abiJSON string
}

// extractDealID extracts a *big.Int deal ID from contract call result data.
func extractDealID(data []interface{}) *big.Int {
	if len(data) > 0 {
		if id, ok := data[0].(*big.Int); ok {
			return id
		}
	}
	return nil
}

// NewHubV2Client creates a hub V2 client for the given contract address.
func NewHubV2Client(caller contract.ContractCaller, address common.Address, chainID int64) *HubV2Client {
	return &HubV2Client{
		caller:  caller,
		address: address,
		chainID: chainID,
		abiJSON: hubV2ABIJSON,
	}
}

// writeMethod executes a state-changing contract call and returns the transaction hash.
func (c *HubV2Client) writeMethod(ctx context.Context, method string, args ...interface{}) (string, error) {
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
func (c *HubV2Client) readMethod(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
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

// DirectSettle transfers tokens directly from buyer to seller without escrow.
func (c *HubV2Client) DirectSettle(ctx context.Context, seller, token common.Address, amount *big.Int, refId [32]byte) (string, error) {
	return c.writeMethod(ctx, MethodDirectSettle, seller, token, amount, refId)
}

// CreateSimpleEscrow creates a simple escrow deal with refId.
func (c *HubV2Client) CreateSimpleEscrow(ctx context.Context, seller, token common.Address, amount, deadline *big.Int, refId [32]byte) (*big.Int, string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  MethodCreateSimpleEscrow,
		Args:    []interface{}{seller, token, amount, deadline, refId},
	})
	if err != nil {
		return nil, "", fmt.Errorf("create simple escrow: %w", err)
	}

	return extractDealID(result.Data), result.TxHash, nil
}

// CreateMilestoneEscrow creates a milestone-based escrow deal with refId.
func (c *HubV2Client) CreateMilestoneEscrow(
	ctx context.Context,
	seller, token common.Address,
	totalAmount *big.Int,
	milestoneAmounts []*big.Int,
	deadline *big.Int,
	refId [32]byte,
) (*big.Int, string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  MethodCreateMilestoneEscrow,
		Args:    []interface{}{seller, token, totalAmount, milestoneAmounts, deadline, refId},
	})
	if err != nil {
		return nil, "", fmt.Errorf("create milestone escrow: %w", err)
	}

	return extractDealID(result.Data), result.TxHash, nil
}

// CreateTeamEscrow creates a team escrow deal with proportional shares and refId.
func (c *HubV2Client) CreateTeamEscrow(
	ctx context.Context,
	members []common.Address,
	token common.Address,
	totalAmount *big.Int,
	shares []*big.Int,
	deadline *big.Int,
	refId [32]byte,
) (*big.Int, string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  MethodCreateTeamEscrow,
		Args:    []interface{}{members, token, totalAmount, shares, deadline, refId},
	})
	if err != nil {
		return nil, "", fmt.Errorf("create team escrow: %w", err)
	}

	return extractDealID(result.Data), result.TxHash, nil
}

// CompleteMilestone marks a milestone as completed on a milestone-type deal.
func (c *HubV2Client) CompleteMilestone(ctx context.Context, dealID *big.Int, index *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodCompleteMilestone, dealID, index)
}

// ReleaseMilestone releases funds for completed milestones.
func (c *HubV2Client) ReleaseMilestone(ctx context.Context, dealID *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodReleaseMilestone, dealID)
}

// Deposit deposits ERC-20 tokens into the V2 escrow.
func (c *HubV2Client) Deposit(ctx context.Context, dealID *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodDeposit, dealID)
}

// Release releases escrow funds to the seller.
func (c *HubV2Client) Release(ctx context.Context, dealID *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodRelease, dealID)
}

// Refund returns escrow funds to the buyer.
func (c *HubV2Client) Refund(ctx context.Context, dealID *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodRefund, dealID)
}

// Dispute raises a dispute on a deal.
func (c *HubV2Client) Dispute(ctx context.Context, dealID *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodDispute, dealID)
}

// ResolveDispute resolves a disputed deal via arbitrator split.
func (c *HubV2Client) ResolveDispute(ctx context.Context, dealID, sellerAmount, buyerAmount *big.Int) (string, error) {
	return c.writeMethod(ctx, MethodResolveDispute, dealID, sellerAmount, buyerAmount)
}

// GetDealV2 reads the on-chain V2 deal state including refId and settler.
func (c *HubV2Client) GetDealV2(ctx context.Context, dealID *big.Int) (*OnChainDealV2, error) {
	data, err := c.readMethod(ctx, MethodGetDeal, dealID)
	if err != nil {
		return nil, err
	}
	return parseDealV2Result(data)
}

// NextDealID reads the next deal ID counter from the V2 hub.
func (c *HubV2Client) NextDealID(ctx context.Context) (*big.Int, error) {
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

// parseDealV2Result converts raw ABI output to OnChainDealV2.
func parseDealV2Result(data []interface{}) (*OnChainDealV2, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty deal v2 result")
	}

	if d, ok := data[0].(struct {
		Buyer    common.Address
		Seller   common.Address
		Token    common.Address
		Amount   *big.Int
		Deadline *big.Int
		Status   uint8
		DealType uint8
		WorkHash [32]byte
		RefId    [32]byte
		Settler  common.Address
	}); ok {
		return &OnChainDealV2{
			OnChainDeal: OnChainDeal{
				Buyer:    d.Buyer,
				Seller:   d.Seller,
				Token:    d.Token,
				Amount:   d.Amount,
				Deadline: d.Deadline,
				Status:   OnChainDealStatus(d.Status),
				WorkHash: d.WorkHash,
			},
			DealType: OnChainDealType(d.DealType),
			RefId:    d.RefId,
			Settler:  d.Settler,
		}, nil
	}

	return nil, fmt.Errorf("unexpected deal v2 result type: %T", data[0])
}
