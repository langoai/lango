package team

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildEscrowTools creates high-level workflow tools that combine team + escrow + budget.
func BuildEscrowTools(coord *Coordinator, escrowEngine *escrow.Engine, budgetEngine *budget.Engine) []*agent.Tool {
	var tools []*agent.Tool

	// 1. team_form_with_budget — combines team formation + escrow creation + budget allocation.
	tools = append(tools, &agent.Tool{
		Name:        "team_form_with_budget",
		Description: "Form a team with automatic escrow and budget allocation in a single step",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name":        map[string]interface{}{"type": "string", "description": "Team name"},
				"goal":        map[string]interface{}{"type": "string", "description": "Team goal/mission"},
				"capability":  map[string]interface{}{"type": "string", "description": "Required capability for workers"},
				"memberCount": map[string]interface{}{"type": "integer", "description": "Number of workers to recruit"},
				"leaderDid":   map[string]interface{}{"type": "string", "description": "DID of the team leader"},
				"budget":      map[string]interface{}{"type": "number", "description": "Total budget in USDC"},
				"milestones": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"description": map[string]interface{}{"type": "string"},
							"amount":      map[string]interface{}{"type": "number", "description": "Amount in USDC"},
						},
					},
					"description": "Milestone definitions (if empty, auto-split evenly among workers)",
				},
			},
			"required": []string{"name", "goal", "capability", "memberCount", "leaderDid", "budget"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			name := toolparam.OptionalString(params, "name", "")
			goal := toolparam.OptionalString(params, "goal", "")
			capability := toolparam.OptionalString(params, "capability", "")
			leaderDID := toolparam.OptionalString(params, "leaderDid", "")
			memberCount := toolparam.OptionalInt(params, "memberCount", 1)
			budgetAmount := toolparam.OptionalFloat64(params, "budget", 0.0)

			if name == "" || capability == "" || leaderDID == "" || budgetAmount <= 0 {
				return nil, fmt.Errorf("missing required parameters or invalid budget")
			}

			// Step 1: Form team.
			teamID := uuid.New().String()
			t, err := coord.FormTeam(ctx, FormTeamRequest{
				TeamID:      teamID,
				Name:        name,
				Goal:        goal,
				LeaderDID:   leaderDID,
				Capability:  capability,
				MemberCount: memberCount,
			})
			if err != nil {
				return nil, fmt.Errorf("form team: %w", err)
			}

			// Set budget on team.
			t.Budget = budgetAmount

			// Collect workers.
			var workers []*Member
			for _, m := range t.Members() {
				if m.Role == RoleWorker {
					workers = append(workers, m)
				}
			}

			// Step 2: Create escrow.
			totalAmount := big.NewInt(int64(budgetAmount * 1_000_000)) // USDC 6 decimals

			// Build milestones.
			var milestones []escrow.MilestoneRequest
			if rawMilestones, ok := params["milestones"].([]interface{}); ok && len(rawMilestones) > 0 {
				milestoneTotal := new(big.Int)
				for _, rm := range rawMilestones {
					ms, _ := rm.(map[string]interface{})
					desc, _ := ms["description"].(string)
					amt := 0.0
					if a, ok := ms["amount"].(float64); ok {
						amt = a
					}
					msAmount := big.NewInt(int64(amt * 1_000_000))
					milestoneTotal.Add(milestoneTotal, msAmount)
					milestones = append(milestones, escrow.MilestoneRequest{
						Description: desc,
						Amount:      msAmount,
					})
				}
				// Adjust total to match milestone sum.
				totalAmount = milestoneTotal
			} else if len(workers) > 0 {
				// Auto-split evenly among workers.
				perWorker := new(big.Int).Div(totalAmount, big.NewInt(int64(len(workers))))
				remainder := new(big.Int).Sub(totalAmount, new(big.Int).Mul(perWorker, big.NewInt(int64(len(workers)))))
				for i, w := range workers {
					amount := new(big.Int).Set(perWorker)
					if i == len(workers)-1 {
						amount.Add(amount, remainder)
					}
					milestones = append(milestones, escrow.MilestoneRequest{
						Description: fmt.Sprintf("Task completion by %s", w.DID),
						Amount:      amount,
					})
				}
			} else {
				// Single milestone for the whole amount.
				milestones = append(milestones, escrow.MilestoneRequest{
					Description: "Team task completion",
					Amount:      totalAmount,
				})
			}

			sellerDID := leaderDID
			if len(workers) > 0 {
				sellerDID = workers[0].DID
			}

			escrowEntry, err := escrowEngine.Create(ctx, escrow.CreateRequest{
				BuyerDID:   leaderDID,
				SellerDID:  sellerDID,
				Amount:     totalAmount,
				Reason:     fmt.Sprintf("Team %s: %s", name, goal),
				TaskID:     teamID,
				Milestones: milestones,
			})
			if err != nil {
				return nil, fmt.Errorf("create escrow: %w", err)
			}

			// Step 3: Allocate budget.
			var budgetID string
			if budgetEngine != nil {
				tb, budgetErr := budgetEngine.Allocate(teamID, totalAmount)
				if budgetErr != nil {
					// Non-fatal: log and continue.
					budgetID = "allocation_failed"
				} else {
					budgetID = tb.TaskID
				}
			}

			// Build members list.
			memberList := make([]map[string]interface{}, 0, len(t.Members()))
			for _, m := range t.Members() {
				memberList = append(memberList, map[string]interface{}{
					"did":  m.DID,
					"name": m.Name,
					"role": string(m.Role),
				})
			}

			return map[string]interface{}{
				"teamId":     teamID,
				"name":       name,
				"goal":       goal,
				"status":     string(t.Status),
				"escrowId":   escrowEntry.ID,
				"budgetId":   budgetID,
				"budget":     budgetAmount,
				"members":    memberList,
				"milestones": len(milestones),
				"createdAt":  t.CreatedAt.Format(time.RFC3339),
			}, nil
		},
	})

	// 2. team_complete_milestone — marks milestone complete and auto-releases if all done.
	tools = append(tools, &agent.Tool{
		Name:        "team_complete_milestone",
		Description: "Mark a team escrow milestone as complete, auto-releases funds when all milestones are done",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"escrowId":    map[string]interface{}{"type": "string", "description": "Escrow ID"},
				"milestoneId": map[string]interface{}{"type": "string", "description": "Milestone ID to complete"},
				"evidence":    map[string]interface{}{"type": "string", "description": "Evidence of milestone completion"},
			},
			"required": []string{"escrowId", "milestoneId"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			escrowID := toolparam.OptionalString(params, "escrowId", "")
			milestoneID := toolparam.OptionalString(params, "milestoneId", "")
			evidence := toolparam.OptionalString(params, "evidence", "")
			if escrowID == "" || milestoneID == "" {
				return nil, fmt.Errorf("missing escrowId or milestoneId")
			}
			if evidence == "" {
				evidence = "manual completion"
			}

			entry, err := escrowEngine.CompleteMilestone(ctx, escrowID, milestoneID, evidence)
			if err != nil {
				return nil, fmt.Errorf("complete milestone: %w", err)
			}

			result := map[string]interface{}{
				"escrowId":            escrowID,
				"milestoneId":         milestoneID,
				"status":              string(entry.Status),
				"completedMilestones": entry.CompletedMilestones(),
				"totalMilestones":     len(entry.Milestones),
				"allCompleted":        entry.AllMilestonesCompleted(),
			}

			// Auto-release if all milestones completed and escrow is in completed state.
			if entry.AllMilestonesCompleted() && entry.Status == escrow.StatusCompleted {
				released, releaseErr := escrowEngine.Release(ctx, escrowID)
				if releaseErr != nil {
					result["releaseError"] = releaseErr.Error()
				} else {
					result["released"] = true
					result["status"] = string(released.Status)
				}
			}

			return result, nil
		},
	})

	return tools
}
