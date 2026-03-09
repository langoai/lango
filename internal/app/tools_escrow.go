package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/escrow/hub"
	"github.com/langoai/lango/internal/wallet"
)

// buildOnChainEscrowTools creates escrow tools with on-chain settlement support.
// The settler parameter is type-asserted at runtime to determine hub vs vault mode.
func buildOnChainEscrowTools(ee *escrow.Engine, settler escrow.SettlementExecutor) []*agent.Tool {
	return []*agent.Tool{
		escrowCreateTool(ee),
		escrowFundTool(ee, settler),
		escrowActivateTool(ee),
		escrowSubmitWorkTool(ee, settler),
		escrowReleaseTool(ee, settler),
		escrowRefundTool(ee, settler),
		escrowDisputeTool(ee, settler),
		escrowResolveTool(ee, settler),
		escrowStatusTool(ee, settler),
		escrowListTool(ee),
	}
}

func escrowCreateTool(ee *escrow.Engine) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_create",
		Description: "Create a new escrow deal between buyer and seller with milestones",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"buyerDid":  map[string]interface{}{"type": "string", "description": "Buyer peer DID"},
				"sellerDid": map[string]interface{}{"type": "string", "description": "Seller peer DID"},
				"amount":    map[string]interface{}{"type": "string", "description": "Total amount in USDC (e.g. '5.00')"},
				"reason":    map[string]interface{}{"type": "string", "description": "Reason for escrow"},
				"milestones": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"description": map[string]interface{}{"type": "string"},
							"amount":      map[string]interface{}{"type": "string"},
						},
					},
					"description": "Milestones with description and amount in USDC",
				},
			},
			"required": []string{"buyerDid", "sellerDid", "amount", "milestones"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			buyerDID, _ := params["buyerDid"].(string)
			sellerDID, _ := params["sellerDid"].(string)
			amtStr, _ := params["amount"].(string)
			reason, _ := params["reason"].(string)

			if buyerDID == "" || sellerDID == "" || amtStr == "" {
				return nil, fmt.Errorf("buyerDid, sellerDid, and amount are required")
			}

			totalAmount, err := wallet.ParseUSDC(amtStr)
			if err != nil {
				return nil, fmt.Errorf("parse amount: %w", err)
			}

			rawMilestones, _ := params["milestones"].([]interface{})
			milestones := make([]escrow.MilestoneRequest, 0, len(rawMilestones))
			for _, rm := range rawMilestones {
				m, ok := rm.(map[string]interface{})
				if !ok {
					continue
				}
				desc, _ := m["description"].(string)
				mAmtStr, _ := m["amount"].(string)
				mAmt, err := wallet.ParseUSDC(mAmtStr)
				if err != nil {
					return nil, fmt.Errorf("parse milestone amount %q: %w", mAmtStr, err)
				}
				milestones = append(milestones, escrow.MilestoneRequest{
					Description: desc,
					Amount:      mAmt,
				})
			}

			entry, err := ee.Create(ctx, escrow.CreateRequest{
				BuyerDID:   buyerDID,
				SellerDID:  sellerDID,
				Amount:     totalAmount,
				Reason:     reason,
				Milestones: milestones,
			})
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"escrowId": entry.ID,
				"status":   string(entry.Status),
				"amount":   wallet.FormatUSDC(entry.TotalAmount),
			}, nil
		},
	}
}

func escrowFundTool(ee *escrow.Engine, settler escrow.SettlementExecutor) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_fund",
		Description: "Fund an escrow with USDC. In on-chain mode, also deposits to the contract.",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID to fund"},
			},
			"required": []string{"escrowId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID, _ := params["escrowId"].(string)
			if escrowID == "" {
				return nil, fmt.Errorf("escrowId is required")
			}

			entry, err := ee.Fund(ctx, escrowID)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"escrowId": entry.ID,
				"status":   string(entry.Status),
				"amount":   wallet.FormatUSDC(entry.TotalAmount),
			}

			// On-chain deposit for hub mode.
			if hs, ok := settler.(*hub.HubSettler); ok {
				if dealID, exists := hs.GetDealID(escrowID); exists {
					txHash, err := hs.HubClient().Deposit(ctx, dealID)
					if err != nil {
						return nil, fmt.Errorf("on-chain deposit: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["dealId"] = dealID.String()
				}
			}

			// On-chain deposit for vault mode.
			if vs, ok := settler.(*hub.VaultSettler); ok {
				if vaultAddr, exists := vs.GetVaultAddress(escrowID); exists {
					vc := vs.VaultClientFor(vaultAddr)
					txHash, err := vc.Deposit(ctx)
					if err != nil {
						return nil, fmt.Errorf("vault deposit: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["vaultAddress"] = vaultAddr.Hex()
				}
			}

			return result, nil
		},
	}
}

func escrowActivateTool(ee *escrow.Engine) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_activate",
		Description: "Activate a funded escrow so work can begin",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID to activate"},
			},
			"required": []string{"escrowId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID, _ := params["escrowId"].(string)
			if escrowID == "" {
				return nil, fmt.Errorf("escrowId is required")
			}

			entry, err := ee.Activate(ctx, escrowID)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"escrowId": entry.ID,
				"status":   string(entry.Status),
			}, nil
		},
	}
}

func escrowSubmitWorkTool(ee *escrow.Engine, settler escrow.SettlementExecutor) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_submit_work",
		Description: "Submit a work hash as proof of completion",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID"},
				"workHash": map[string]interface{}{"type": "string", "description": "Work proof hash (will be SHA-256 hashed for on-chain submission)"},
			},
			"required": []string{"escrowId", "workHash"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID, _ := params["escrowId"].(string)
			workHashStr, _ := params["workHash"].(string)
			if escrowID == "" || workHashStr == "" {
				return nil, fmt.Errorf("escrowId and workHash are required")
			}

			// Verify the escrow exists and is active.
			entry, err := ee.Get(escrowID)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"escrowId": entry.ID,
				"status":   string(entry.Status),
				"workHash": workHashStr,
			}

			workHash := sha256.Sum256([]byte(workHashStr))

			// On-chain submit for hub mode.
			if hs, ok := settler.(*hub.HubSettler); ok {
				if dealID, exists := hs.GetDealID(escrowID); exists {
					txHash, err := hs.HubClient().SubmitWork(ctx, dealID, workHash)
					if err != nil {
						return nil, fmt.Errorf("on-chain submit work: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["dealId"] = dealID.String()
				}
			}

			// On-chain submit for vault mode.
			if vs, ok := settler.(*hub.VaultSettler); ok {
				if vaultAddr, exists := vs.GetVaultAddress(escrowID); exists {
					vc := vs.VaultClientFor(vaultAddr)
					txHash, err := vc.SubmitWork(ctx, workHash)
					if err != nil {
						return nil, fmt.Errorf("vault submit work: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["vaultAddress"] = vaultAddr.Hex()
				}
			}

			return result, nil
		},
	}
}

func escrowReleaseTool(ee *escrow.Engine, settler escrow.SettlementExecutor) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_release",
		Description: "Release escrow funds to the seller",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID to release"},
			},
			"required": []string{"escrowId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID, _ := params["escrowId"].(string)
			if escrowID == "" {
				return nil, fmt.Errorf("escrowId is required")
			}

			entry, err := ee.Release(ctx, escrowID)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"escrowId": entry.ID,
				"status":   string(entry.Status),
			}

			// On-chain release for hub mode.
			if hs, ok := settler.(*hub.HubSettler); ok {
				if dealID, exists := hs.GetDealID(escrowID); exists {
					txHash, err := hs.HubClient().Release(ctx, dealID)
					if err != nil {
						return nil, fmt.Errorf("on-chain release: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["dealId"] = dealID.String()
				}
			}

			// On-chain release for vault mode.
			if vs, ok := settler.(*hub.VaultSettler); ok {
				if vaultAddr, exists := vs.GetVaultAddress(escrowID); exists {
					vc := vs.VaultClientFor(vaultAddr)
					txHash, err := vc.Release(ctx)
					if err != nil {
						return nil, fmt.Errorf("vault release: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["vaultAddress"] = vaultAddr.Hex()
				}
			}

			return result, nil
		},
	}
}

func escrowRefundTool(ee *escrow.Engine, settler escrow.SettlementExecutor) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_refund",
		Description: "Refund escrow funds to the buyer",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID to refund"},
			},
			"required": []string{"escrowId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID, _ := params["escrowId"].(string)
			if escrowID == "" {
				return nil, fmt.Errorf("escrowId is required")
			}

			entry, err := ee.Refund(ctx, escrowID)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"escrowId": entry.ID,
				"status":   string(entry.Status),
			}

			// On-chain refund for hub mode.
			if hs, ok := settler.(*hub.HubSettler); ok {
				if dealID, exists := hs.GetDealID(escrowID); exists {
					txHash, err := hs.HubClient().Refund(ctx, dealID)
					if err != nil {
						return nil, fmt.Errorf("on-chain refund: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["dealId"] = dealID.String()
				}
			}

			// On-chain refund for vault mode.
			if vs, ok := settler.(*hub.VaultSettler); ok {
				if vaultAddr, exists := vs.GetVaultAddress(escrowID); exists {
					vc := vs.VaultClientFor(vaultAddr)
					txHash, err := vc.Refund(ctx)
					if err != nil {
						return nil, fmt.Errorf("vault refund: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["vaultAddress"] = vaultAddr.Hex()
				}
			}

			return result, nil
		},
	}
}

func escrowDisputeTool(ee *escrow.Engine, settler escrow.SettlementExecutor) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_dispute",
		Description: "Raise a dispute on an escrow",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID to dispute"},
				"note":     map[string]interface{}{"type": "string", "description": "Dispute description"},
			},
			"required": []string{"escrowId", "note"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID, _ := params["escrowId"].(string)
			note, _ := params["note"].(string)
			if escrowID == "" || note == "" {
				return nil, fmt.Errorf("escrowId and note are required")
			}

			entry, err := ee.Dispute(ctx, escrowID, note)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"escrowId": entry.ID,
				"status":   string(entry.Status),
			}

			// On-chain dispute for hub mode.
			if hs, ok := settler.(*hub.HubSettler); ok {
				if dealID, exists := hs.GetDealID(escrowID); exists {
					txHash, err := hs.HubClient().Dispute(ctx, dealID)
					if err != nil {
						return nil, fmt.Errorf("on-chain dispute: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["dealId"] = dealID.String()
				}
			}

			// On-chain dispute for vault mode.
			if vs, ok := settler.(*hub.VaultSettler); ok {
				if vaultAddr, exists := vs.GetVaultAddress(escrowID); exists {
					vc := vs.VaultClientFor(vaultAddr)
					txHash, err := vc.Dispute(ctx)
					if err != nil {
						return nil, fmt.Errorf("vault dispute: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["vaultAddress"] = vaultAddr.Hex()
				}
			}

			return result, nil
		},
	}
}

func escrowResolveTool(ee *escrow.Engine, settler escrow.SettlementExecutor) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_resolve",
		Description: "Resolve a disputed escrow as arbitrator. Specify favor and seller percentage.",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId":      map[string]interface{}{"type": "string", "description": "Escrow ID to resolve"},
				"favor":         map[string]interface{}{"type": "string", "description": "Which party is favored", "enum": []string{"buyer", "seller"}},
				"sellerPercent": map[string]interface{}{"type": "number", "description": "Percentage of funds to seller (0-100)"},
			},
			"required": []string{"escrowId", "favor", "sellerPercent"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID, _ := params["escrowId"].(string)
			favor, _ := params["favor"].(string)
			sellerPctFloat, _ := params["sellerPercent"].(float64)
			if escrowID == "" || favor == "" {
				return nil, fmt.Errorf("escrowId and favor are required")
			}
			if sellerPctFloat < 0 || sellerPctFloat > 100 {
				return nil, fmt.Errorf("sellerPercent must be between 0 and 100")
			}

			entry, err := ee.Get(escrowID)
			if err != nil {
				return nil, err
			}

			sellerFavor := favor == "seller"
			sellerPct := int64(sellerPctFloat)
			sellerAmount := new(big.Int).Mul(entry.TotalAmount, big.NewInt(sellerPct))
			sellerAmount.Div(sellerAmount, big.NewInt(100))
			buyerAmount := new(big.Int).Sub(entry.TotalAmount, sellerAmount)

			result := map[string]interface{}{
				"escrowId":     entry.ID,
				"favor":        favor,
				"sellerAmount": wallet.FormatUSDC(sellerAmount),
				"buyerAmount":  wallet.FormatUSDC(buyerAmount),
			}

			// On-chain resolve for hub mode.
			if hs, ok := settler.(*hub.HubSettler); ok {
				if dealID, exists := hs.GetDealID(escrowID); exists {
					txHash, err := hs.HubClient().ResolveDispute(ctx, dealID, sellerFavor, sellerAmount, buyerAmount)
					if err != nil {
						return nil, fmt.Errorf("on-chain resolve: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["dealId"] = dealID.String()
				}
			}

			// On-chain resolve for vault mode.
			if vs, ok := settler.(*hub.VaultSettler); ok {
				if vaultAddr, exists := vs.GetVaultAddress(escrowID); exists {
					vc := vs.VaultClientFor(vaultAddr)
					txHash, err := vc.Resolve(ctx, sellerFavor, sellerAmount, buyerAmount)
					if err != nil {
						return nil, fmt.Errorf("vault resolve: %w", err)
					}
					result["onChainTxHash"] = txHash
					result["vaultAddress"] = vaultAddr.Hex()
				}
			}

			return result, nil
		},
	}
}

func escrowStatusTool(ee *escrow.Engine, settler escrow.SettlementExecutor) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_status",
		Description: "Get detailed escrow status including on-chain state if available",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID"},
			},
			"required": []string{"escrowId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID, _ := params["escrowId"].(string)
			if escrowID == "" {
				return nil, fmt.Errorf("escrowId is required")
			}

			entry, err := ee.Get(escrowID)
			if err != nil {
				return nil, err
			}

			milestones := make([]map[string]interface{}, len(entry.Milestones))
			for i, m := range entry.Milestones {
				milestones[i] = map[string]interface{}{
					"id":          m.ID,
					"description": m.Description,
					"amount":      wallet.FormatUSDC(m.Amount),
					"status":      string(m.Status),
				}
				if m.CompletedAt != nil {
					milestones[i]["completedAt"] = m.CompletedAt.Format("2006-01-02T15:04:05Z")
				}
			}

			result := map[string]interface{}{
				"escrowId":   entry.ID,
				"buyerDid":   entry.BuyerDID,
				"sellerDid":  entry.SellerDID,
				"amount":     wallet.FormatUSDC(entry.TotalAmount),
				"status":     string(entry.Status),
				"reason":     entry.Reason,
				"milestones": milestones,
				"expiresAt":  entry.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			}

			// Enrich with on-chain state for hub mode.
			if hs, ok := settler.(*hub.HubSettler); ok {
				if dealID, exists := hs.GetDealID(escrowID); exists {
					result["dealId"] = dealID.String()
					deal, err := hs.HubClient().GetDeal(ctx, dealID)
					if err == nil {
						result["onChainStatus"] = deal.Status.String()
						result["onChainAmount"] = deal.Amount.String()
					}
				}
			}

			// Enrich with on-chain state for vault mode.
			if vs, ok := settler.(*hub.VaultSettler); ok {
				if vaultAddr, exists := vs.GetVaultAddress(escrowID); exists {
					result["vaultAddress"] = vaultAddr.Hex()
					vc := vs.VaultClientFor(vaultAddr)
					status, err := vc.Status(ctx)
					if err == nil {
						result["onChainStatus"] = status.String()
					}
					amount, err := vc.Amount(ctx)
					if err == nil {
						result["onChainAmount"] = amount.String()
					}
				}
			}

			return result, nil
		},
	}
}

func escrowListTool(ee *escrow.Engine) *agent.Tool {
	return &agent.Tool{
		Name:        "escrow_list",
		Description: "List all escrows with optional filter by status or peer",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filter":  map[string]interface{}{"type": "string", "description": "Filter by status: all, active, disputed", "enum": []string{"all", "active", "disputed"}},
				"peerDid": map[string]interface{}{"type": "string", "description": "Filter by peer DID (buyer or seller)"},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			filter, _ := params["filter"].(string)
			peerDID, _ := params["peerDid"].(string)

			var entries []*escrow.EscrowEntry
			if peerDID != "" {
				entries = ee.ListByPeer(peerDID)
			} else {
				entries = ee.List()
			}

			// Apply status filter.
			if filter == "active" || filter == "disputed" {
				filtered := make([]*escrow.EscrowEntry, 0, len(entries))
				for _, e := range entries {
					if filter == "active" && (e.Status == escrow.StatusActive || e.Status == escrow.StatusFunded) {
						filtered = append(filtered, e)
					}
					if filter == "disputed" && e.Status == escrow.StatusDisputed {
						filtered = append(filtered, e)
					}
				}
				entries = filtered
			}

			items := make([]map[string]interface{}, len(entries))
			for i, e := range entries {
				items[i] = map[string]interface{}{
					"escrowId":  e.ID,
					"buyerDid":  e.BuyerDID,
					"sellerDid": e.SellerDID,
					"amount":    wallet.FormatUSDC(e.TotalAmount),
					"status":    string(e.Status),
					"reason":    e.Reason,
				}
			}

			return map[string]interface{}{
				"count":   len(items),
				"escrows": items,
			}, nil
		},
	}
}
