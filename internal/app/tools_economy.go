package app

import (
	"context"
	"fmt"
	"math/big"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/negotiation"
	"github.com/langoai/lango/internal/economy/pricing"
	"github.com/langoai/lango/internal/economy/risk"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/langoai/lango/internal/wallet"
)

// buildEconomyTools creates economy layer tools from engine components.
func buildEconomyTools(ec *economyComponents) []*agent.Tool {
	tools := make([]*agent.Tool, 0, 12)

	if ec.budgetEngine != nil {
		tools = append(tools, buildBudgetTools(ec.budgetEngine)...)
	}
	if ec.riskEngine != nil {
		tools = append(tools, buildRiskTools(ec.riskEngine)...)
	}
	if ec.negotiationEngine != nil {
		tools = append(tools, buildNegotiationTools(ec.negotiationEngine)...)
	}
	if ec.escrowEngine != nil {
		tools = append(tools, buildEscrowTools(ec.escrowEngine)...)
	}
	if ec.pricingEngine != nil {
		tools = append(tools, buildPricingTools(ec.pricingEngine)...)
	}

	return tools
}

func buildBudgetTools(be *budget.Engine) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "economy_budget_allocate",
			Description: "Allocate a spending budget for a task (amount in USDC, e.g. '5.00')",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"taskId": map[string]interface{}{"type": "string", "description": "Unique task identifier"},
					"amount": map[string]interface{}{"type": "string", "description": "Budget in USDC (e.g. '5.00'). Omit for default max."},
				},
				"required": []string{"taskId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				taskID, err := toolparam.RequireString(params, "taskId")
				if err != nil {
					return nil, err
				}
				var total *big.Int
				if amtStr := toolparam.OptionalString(params, "amount", ""); amtStr != "" {
					parsed, err := wallet.ParseUSDC(amtStr)
					if err != nil {
						return nil, fmt.Errorf("parse amount %q: %w", amtStr, err)
					}
					total = parsed
				}
				tb, err := be.Allocate(taskID, total)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"taskId":      tb.TaskID,
					"totalBudget": tb.TotalBudget.String(),
					"status":      string(tb.Status),
				}, nil
			},
		},
		{
			Name:        "economy_budget_status",
			Description: "Check budget status for a task",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"taskId": map[string]interface{}{"type": "string", "description": "Task identifier"},
				},
				"required": []string{"taskId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				taskID, err := toolparam.RequireString(params, "taskId")
				if err != nil {
					return nil, err
				}
				rate, err := be.BurnRate(taskID)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"taskId":   taskID,
					"burnRate": rate.String(),
				}, nil
			},
		},
		{
			Name:        "economy_budget_close",
			Description: "Close a task budget and get final report",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"taskId": map[string]interface{}{"type": "string", "description": "Task identifier"},
				},
				"required": []string{"taskId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				taskID, err := toolparam.RequireString(params, "taskId")
				if err != nil {
					return nil, err
				}
				report, err := be.Close(taskID)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"taskId":     report.TaskID,
					"totalSpent": report.TotalSpent.String(),
					"entries":    report.EntryCount,
					"status":     string(report.Status),
				}, nil
			},
		},
	}
}

func buildRiskTools(re *risk.Engine) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "economy_risk_assess",
			Description: "Assess risk for a transaction with a peer",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peerDid":       map[string]interface{}{"type": "string", "description": "Peer DID"},
					"amount":        map[string]interface{}{"type": "string", "description": "Transaction amount in USDC (e.g. '1.00')"},
					"verifiability": map[string]interface{}{"type": "string", "description": "Output verifiability: high, medium, low", "enum": []string{"high", "medium", "low"}},
				},
				"required": []string{"peerDid", "amount"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peerDid")
				if err != nil {
					return nil, err
				}
				amtStr, err := toolparam.RequireString(params, "amount")
				if err != nil {
					return nil, err
				}
				amount, err := wallet.ParseUSDC(amtStr)
				if err != nil {
					return nil, fmt.Errorf("parse amount: %w", err)
				}
				v := risk.VerifiabilityMedium
				if vs := toolparam.OptionalString(params, "verifiability", ""); vs != "" {
					v = risk.Verifiability(vs)
				}
				assessment, err := re.Assess(ctx, peerDID, amount, v)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"riskLevel":   string(assessment.RiskLevel),
					"riskScore":   assessment.RiskScore,
					"strategy":    string(assessment.Strategy),
					"trustScore":  assessment.TrustScore,
					"explanation": assessment.Explanation,
				}, nil
			},
		},
	}
}

func buildNegotiationTools(ne *negotiation.Engine) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "economy_negotiate",
			Description: "Start a price negotiation with a peer",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peerDid":  map[string]interface{}{"type": "string", "description": "Responder peer DID"},
					"toolName": map[string]interface{}{"type": "string", "description": "Tool to negotiate price for"},
					"price":    map[string]interface{}{"type": "string", "description": "Proposed price in USDC (e.g. '1.00')"},
				},
				"required": []string{"peerDid", "toolName", "price"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peerDid")
				if err != nil {
					return nil, err
				}
				toolName, err := toolparam.RequireString(params, "toolName")
				if err != nil {
					return nil, err
				}
				priceStr, err := toolparam.RequireString(params, "price")
				if err != nil {
					return nil, err
				}
				price, err := wallet.ParseUSDC(priceStr)
				if err != nil {
					return nil, fmt.Errorf("parse price: %w", err)
				}
				terms := negotiation.Terms{
					ToolName: toolName,
					Price:    price,
					Currency: "USDC",
				}
				sess, err := ne.Propose(ctx, "local", peerDID, terms)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"sessionId": sess.ID,
					"phase":     string(sess.Phase),
					"round":     sess.Round,
				}, nil
			},
		},
		{
			Name:        "economy_negotiate_status",
			Description: "Check the status of a negotiation session",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sessionId": map[string]interface{}{"type": "string", "description": "Negotiation session ID"},
				},
				"required": []string{"sessionId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionID, err := toolparam.RequireString(params, "sessionId")
				if err != nil {
					return nil, err
				}
				sess, err := ne.Get(sessionID)
				if err != nil {
					return nil, err
				}
				result := map[string]interface{}{
					"sessionId":    sess.ID,
					"phase":        string(sess.Phase),
					"round":        sess.Round,
					"maxRounds":    sess.MaxRounds,
					"initiatorDid": sess.InitiatorDID,
					"responderDid": sess.ResponderDID,
				}
				if sess.CurrentTerms != nil {
					result["toolName"] = sess.CurrentTerms.ToolName
					result["price"] = sess.CurrentTerms.Price.String()
				}
				return result, nil
			},
		},
	}
}

func buildEscrowTools(ee *escrow.Engine) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "economy_escrow_create",
			Description: "Create a milestone-based escrow between buyer and seller",
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
						"description": "Milestones with description and amount",
					},
				},
				"required": []string{"buyerDid", "sellerDid", "amount", "milestones"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				buyerDID := toolparam.OptionalString(params, "buyerDid", "")
				sellerDID := toolparam.OptionalString(params, "sellerDid", "")
				amtStr := toolparam.OptionalString(params, "amount", "")
				reasonStr := toolparam.OptionalString(params, "reason", "")

				totalAmount, err := wallet.ParseUSDC(amtStr)
				if err != nil {
					return nil, fmt.Errorf("parse amount: %w", err)
				}

				rawMilestones, _ := params["milestones"].([]interface{})
				var milestones []escrow.MilestoneRequest
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
					Reason:     reasonStr,
					Milestones: milestones,
				})
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"escrowId": entry.ID,
					"status":   string(entry.Status),
					"amount":   entry.TotalAmount.String(),
				}, nil
			},
		},
		{
			Name:        "economy_escrow_milestone",
			Description: "Complete a milestone in an escrow",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"escrowId":    map[string]interface{}{"type": "string", "description": "Escrow ID"},
					"milestoneId": map[string]interface{}{"type": "string", "description": "Milestone ID"},
					"evidence":    map[string]interface{}{"type": "string", "description": "Evidence of completion"},
				},
				"required": []string{"escrowId", "milestoneId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				escrowID := toolparam.OptionalString(params, "escrowId", "")
				milestoneID := toolparam.OptionalString(params, "milestoneId", "")
				evidence := toolparam.OptionalString(params, "evidence", "")
				entry, err := ee.CompleteMilestone(ctx, escrowID, milestoneID, evidence)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"escrowId":            entry.ID,
					"status":              string(entry.Status),
					"completedMilestones": entry.CompletedMilestones(),
					"totalMilestones":     len(entry.Milestones),
				}, nil
			},
		},
		{
			Name:        "economy_escrow_status",
			Description: "Check escrow status",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID"},
				},
				"required": []string{"escrowId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				escrowID := toolparam.OptionalString(params, "escrowId", "")
				entry, err := ee.Get(escrowID)
				if err != nil {
					return nil, err
				}
				milestones := make([]map[string]interface{}, len(entry.Milestones))
				for i, m := range entry.Milestones {
					milestones[i] = map[string]interface{}{
						"id":          m.ID,
						"description": m.Description,
						"amount":      m.Amount.String(),
						"status":      string(m.Status),
					}
				}
				return map[string]interface{}{
					"escrowId":   entry.ID,
					"buyerDid":   entry.BuyerDID,
					"sellerDid":  entry.SellerDID,
					"amount":     entry.TotalAmount.String(),
					"status":     string(entry.Status),
					"milestones": milestones,
				}, nil
			},
		},
		{
			Name:        "economy_escrow_release",
			Description: "Release escrow funds to seller",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID"},
				},
				"required": []string{"escrowId"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				escrowID := toolparam.OptionalString(params, "escrowId", "")
				entry, err := ee.Release(ctx, escrowID)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"escrowId": entry.ID,
					"status":   string(entry.Status),
				}, nil
			},
		},
		{
			Name:        "economy_escrow_dispute",
			Description: "Raise a dispute on an escrow",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"escrowId": map[string]interface{}{"type": "string", "description": "Escrow ID"},
					"note":     map[string]interface{}{"type": "string", "description": "Dispute description"},
				},
				"required": []string{"escrowId", "note"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				escrowID := toolparam.OptionalString(params, "escrowId", "")
				note := toolparam.OptionalString(params, "note", "")
				entry, err := ee.Dispute(ctx, escrowID, note)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"escrowId": entry.ID,
					"status":   string(entry.Status),
				}, nil
			},
		},
	}
}

func buildPricingTools(pe *pricing.Engine) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "economy_price_quote",
			Description: "Get a price quote for a tool, optionally with peer-specific discounts",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"toolName": map[string]interface{}{"type": "string", "description": "Tool name to quote"},
					"peerDid":  map[string]interface{}{"type": "string", "description": "Optional peer DID for trust discounts"},
				},
				"required": []string{"toolName"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				toolName, err := toolparam.RequireString(params, "toolName")
				if err != nil {
					return nil, err
				}
				peerDID := toolparam.OptionalString(params, "peerDid", "")
				quote, err := pe.Quote(ctx, toolName, peerDID)
				if err != nil {
					return nil, err
				}
				result := map[string]interface{}{
					"toolName": quote.ToolName,
					"isFree":   quote.IsFree,
				}
				if !quote.IsFree {
					result["basePrice"] = quote.BasePrice.String()
					result["finalPrice"] = quote.FinalPrice.String()
					result["currency"] = quote.Currency
				}
				return result, nil
			},
		},
	}
}
