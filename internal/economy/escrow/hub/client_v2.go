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

// DirectSettle transfers tokens directly from buyer to seller without escrow.
func (c *HubV2Client) DirectSettle(ctx context.Context, seller, token common.Address, amount *big.Int, refId [32]byte) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "directSettle",
		Args:    []interface{}{seller, token, amount, refId},
	})
	if err != nil {
		return "", fmt.Errorf("direct settle: %w", err)
	}
	return result.TxHash, nil
}

// CreateSimpleEscrow creates a simple escrow deal with refId.
func (c *HubV2Client) CreateSimpleEscrow(ctx context.Context, seller, token common.Address, amount, deadline *big.Int, refId [32]byte) (*big.Int, string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "createSimpleEscrow",
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
		Method:  "createMilestoneEscrow",
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
		Method:  "createTeamEscrow",
		Args:    []interface{}{members, token, totalAmount, shares, deadline, refId},
	})
	if err != nil {
		return nil, "", fmt.Errorf("create team escrow: %w", err)
	}

	return extractDealID(result.Data), result.TxHash, nil
}

// CompleteMilestone marks a milestone as completed on a milestone-type deal.
func (c *HubV2Client) CompleteMilestone(ctx context.Context, dealID *big.Int, index *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "completeMilestone",
		Args:    []interface{}{dealID, index},
	})
	if err != nil {
		return "", fmt.Errorf("complete milestone deal %s idx %s: %w", dealID, index, err)
	}
	return result.TxHash, nil
}

// ReleaseMilestone releases funds for completed milestones.
func (c *HubV2Client) ReleaseMilestone(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "releaseMilestone",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("release milestone deal %s: %w", dealID, err)
	}
	return result.TxHash, nil
}

// Deposit deposits ERC-20 tokens into the V2 escrow.
func (c *HubV2Client) Deposit(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "deposit",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("deposit deal %s: %w", dealID, err)
	}
	return result.TxHash, nil
}

// Release releases escrow funds to the seller.
func (c *HubV2Client) Release(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "release",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("release deal %s: %w", dealID, err)
	}
	return result.TxHash, nil
}

// Refund returns escrow funds to the buyer.
func (c *HubV2Client) Refund(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "refund",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("refund deal %s: %w", dealID, err)
	}
	return result.TxHash, nil
}

// Dispute raises a dispute on a deal.
func (c *HubV2Client) Dispute(ctx context.Context, dealID *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "dispute",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return "", fmt.Errorf("dispute deal %s: %w", dealID, err)
	}
	return result.TxHash, nil
}

// ResolveDispute resolves a disputed deal via arbitrator split.
func (c *HubV2Client) ResolveDispute(ctx context.Context, dealID, sellerAmount, buyerAmount *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "resolveDispute",
		Args:    []interface{}{dealID, sellerAmount, buyerAmount},
	})
	if err != nil {
		return "", fmt.Errorf("resolve dispute deal %s: %w", dealID, err)
	}
	return result.TxHash, nil
}

// GetDealV2 reads the on-chain V2 deal state including refId and settler.
func (c *HubV2Client) GetDealV2(ctx context.Context, dealID *big.Int) (*OnChainDealV2, error) {
	result, err := c.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "getDeal",
		Args:    []interface{}{dealID},
	})
	if err != nil {
		return nil, fmt.Errorf("get deal v2 %s: %w", dealID, err)
	}
	return parseDealV2Result(result.Data)
}

// NextDealID reads the next deal ID counter from the V2 hub.
func (c *HubV2Client) NextDealID(ctx context.Context) (*big.Int, error) {
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
