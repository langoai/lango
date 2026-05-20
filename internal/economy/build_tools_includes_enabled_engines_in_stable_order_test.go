package economy

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/negotiation"
	"github.com/langoai/lango/internal/economy/pricing"
	"github.com/langoai/lango/internal/economy/risk"
	"github.com/langoai/lango/internal/wallet"
)

func TestBuildToolsIncludesEnabledEnginesInStableOrder(t *testing.T) {
	t.Parallel()

	be := newBuildToolsIncludesEnabledEnginesInStableOrderBudgetEngine(t)
	re := newBuildToolsIncludesEnabledEnginesInStableOrderRiskEngine(t, 0.9)
	ne := negotiation.New(config.NegotiationConfig{MaxRounds: 3})
	ee := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())
	pe := newBuildToolsIncludesEnabledEnginesInStableOrderPricingEngine(t)

	tools := BuildTools(be, re, ne, ee, pe)

	var names []string
	for _, tool := range tools {
		names = append(names, tool.Name)
		assert.Equal(t, "economy", tool.Capability.Category)
	}
	assert.Equal(t, []string{
		"economy_budget_allocate",
		"economy_budget_status",
		"economy_budget_close",
		"economy_risk_assess",
		"economy_negotiate",
		"economy_negotiate_status",
		"economy_escrow_create",
		"economy_escrow_milestone",
		"economy_escrow_status",
		"economy_escrow_release",
		"economy_escrow_dispute",
		"economy_price_quote",
	}, names)
}

func TestBudgetToolHandlersUseDefaultsAndReturnReports(t *testing.T) {
	t.Parallel()

	tools := buildBudgetTools(newBuildToolsIncludesEnabledEnginesInStableOrderBudgetEngine(t))

	allocate := findEconomyTool(tools, "economy_budget_allocate")
	require.NotNil(t, allocate)
	got, err := allocate.Handler(context.Background(), map[string]interface{}{"taskId": "task-default"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"taskId":      "task-default",
		"totalBudget": "2500000",
		"status":      "active",
	}, got)

	status := findEconomyTool(tools, "economy_budget_status")
	require.NotNil(t, status)
	got, err = status.Handler(context.Background(), map[string]interface{}{"taskId": "task-default"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"taskId":   "task-default",
		"burnRate": "0",
	}, got)

	closeTool := findEconomyTool(tools, "economy_budget_close")
	require.NotNil(t, closeTool)
	got, err = closeTool.Handler(context.Background(), map[string]interface{}{"taskId": "task-default"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"taskId":     "task-default",
		"totalSpent": "0",
		"entries":    0,
		"status":     "closed",
	}, got)
}

func TestRiskAndNegotiationToolHandlersReturnStructuredResults(t *testing.T) {
	t.Parallel()

	riskTool := findEconomyTool(buildRiskTools(newBuildToolsIncludesEnabledEnginesInStableOrderRiskEngine(t, 0.95)), "economy_risk_assess")
	require.NotNil(t, riskTool)
	got, err := riskTool.Handler(context.Background(), map[string]interface{}{
		"peerDid": "did:lango:trusted",
		"amount":  "1.00",
	})
	require.NoError(t, err)
	riskResult := got.(map[string]interface{})
	assert.Equal(t, "direct_pay", riskResult["strategy"])
	assert.Equal(t, 0.95, riskResult["trustScore"])
	assert.Contains(t, riskResult["explanation"], "peer trust is high")

	negotiationTools := buildNegotiationTools(negotiation.New(config.NegotiationConfig{MaxRounds: 4}))
	negotiate := findEconomyTool(negotiationTools, "economy_negotiate")
	require.NotNil(t, negotiate)
	got, err = negotiate.Handler(context.Background(), map[string]interface{}{
		"peerDid":  "did:lango:seller",
		"toolName": "code_review",
		"price":    "1.25",
	})
	require.NoError(t, err)
	negotiationResult := got.(map[string]interface{})
	sessionID := negotiationResult["sessionId"].(string)
	assert.NotEmpty(t, sessionID)
	assert.Equal(t, "proposed", negotiationResult["phase"])
	assert.Equal(t, 1, negotiationResult["round"])

	status := findEconomyTool(negotiationTools, "economy_negotiate_status")
	require.NotNil(t, status)
	got, err = status.Handler(context.Background(), map[string]interface{}{"sessionId": sessionID})
	require.NoError(t, err)
	statusResult := got.(map[string]interface{})
	assert.Equal(t, sessionID, statusResult["sessionId"])
	assert.Equal(t, "code_review", statusResult["toolName"])
	assert.Equal(t, "1250000", statusResult["price"])
	assert.Equal(t, 4, statusResult["maxRounds"])
}

func TestEscrowCreateSkipsMalformedMilestonesAndStatusExpandsMilestones(t *testing.T) {
	t.Parallel()

	tools := buildEscrowTools(
		escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig()),
	)
	create := findEconomyTool(tools, "economy_escrow_create")
	require.NotNil(t, create)

	got, err := create.Handler(context.Background(), map[string]interface{}{
		"buyerDid":  "did:lango:buyer",
		"sellerDid": "did:lango:seller",
		"amount":    "1.00",
		"reason":    "milestone delivery",
		"milestones": []interface{}{
			"ignore malformed milestone",
			map[string]interface{}{"description": "draft", "amount": "1.00"},
		},
	})
	require.NoError(t, err)
	createResult := got.(map[string]interface{})
	escrowID := createResult["escrowId"].(string)
	assert.NotEmpty(t, escrowID)
	assert.Equal(t, "pending", createResult["status"])
	assert.Equal(t, "1000000", createResult["amount"])

	status := findEconomyTool(tools, "economy_escrow_status")
	require.NotNil(t, status)
	got, err = status.Handler(context.Background(), map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	statusResult := got.(map[string]interface{})
	assert.Equal(t, escrowID, statusResult["escrowId"])
	assert.Equal(t, "did:lango:buyer", statusResult["buyerDid"])
	assert.Equal(t, "did:lango:seller", statusResult["sellerDid"])
	assert.Equal(t, "pending", statusResult["status"])
	milestones := statusResult["milestones"].([]map[string]interface{})
	require.Len(t, milestones, 1)
	assert.Equal(t, "draft", milestones[0]["description"])
	assert.Equal(t, "1000000", milestones[0]["amount"])
	assert.Equal(t, "pending", milestones[0]["status"])
}

func TestPricingToolOmitsPriceFieldsForFreeQuotes(t *testing.T) {
	t.Parallel()

	priceTool := findEconomyTool(buildPricingTools(newBuildToolsIncludesEnabledEnginesInStableOrderPricingEngine(t)), "economy_price_quote")
	require.NotNil(t, priceTool)

	got, err := priceTool.Handler(context.Background(), map[string]interface{}{"toolName": "free_tool"})
	require.NoError(t, err)
	freeResult := got.(map[string]interface{})
	assert.Equal(t, "free_tool", freeResult["toolName"])
	assert.Equal(t, true, freeResult["isFree"])
	assert.NotContains(t, freeResult, "basePrice")
	assert.NotContains(t, freeResult, "finalPrice")
	assert.NotContains(t, freeResult, "currency")

	got, err = priceTool.Handler(context.Background(), map[string]interface{}{
		"toolName": "paid_tool",
		"peerDid":  "did:lango:trusted",
	})
	require.NoError(t, err)
	paidResult := got.(map[string]interface{})
	assert.Equal(t, "paid_tool", paidResult["toolName"])
	assert.Equal(t, false, paidResult["isFree"])
	assert.Equal(t, "2000000", paidResult["basePrice"])
	assert.Equal(t, "2000000", paidResult["finalPrice"])
	assert.Equal(t, "USDC", paidResult["currency"])
}

func newBuildToolsIncludesEnabledEnginesInStableOrderBudgetEngine(t *testing.T) *budget.Engine {
	t.Helper()

	engine, err := budget.NewEngine(
		budget.NewStore(),
		config.BudgetConfig{DefaultMax: "2.50"},
	)
	require.NoError(t, err)
	return engine
}

func newBuildToolsIncludesEnabledEnginesInStableOrderRiskEngine(t *testing.T, trustScore float64) *risk.Engine {
	t.Helper()

	engine, err := risk.New(config.RiskConfig{}, func(context.Context, string) (float64, error) {
		return trustScore, nil
	})
	require.NoError(t, err)
	return engine
}

func newBuildToolsIncludesEnabledEnginesInStableOrderPricingEngine(t *testing.T) *pricing.Engine {
	t.Helper()

	engine, err := pricing.New(config.DynamicPricingConfig{})
	require.NoError(t, err)
	engine.SetBasePrice("paid_tool", buildToolsIncludesEnabledEnginesInStableOrderUSDC(t, "2.00"))
	return engine
}

func buildToolsIncludesEnabledEnginesInStableOrderUSDC(t *testing.T, amount string) *big.Int {
	t.Helper()

	value, err := wallet.ParseUSDC(amount)
	require.NoError(t, err)
	return value
}
