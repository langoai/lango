package negotiation

import (
	"context"
	"math/big"
)

// AutoStrategy generates counter-offers automatically.
type AutoStrategy struct {
	basePrice   *big.Int
	maxDiscount float64 // max discount from base price (0-1)
}

// NewAutoStrategy creates a new auto-strategy.
func NewAutoStrategy(basePrice *big.Int, maxDiscount float64) *AutoStrategy {
	return &AutoStrategy{
		basePrice:   new(big.Int).Set(basePrice),
		maxDiscount: maxDiscount,
	}
}

// GenerateCounter produces a counter-offer given a proposal.
// Strategy: meet halfway between current offer and base price, but never go
// below (1-maxDiscount)*basePrice.
func (s *AutoStrategy) GenerateCounter(proposed *big.Int, round int, maxRounds int) *big.Int {
	// floor = basePrice * (1 - maxDiscount)
	floorBps := int64((1.0 - s.maxDiscount) * 10000)
	floor := new(big.Int).Mul(s.basePrice, big.NewInt(floorBps))
	floor.Div(floor, big.NewInt(10000))

	// midpoint = (proposed + basePrice) / 2
	midpoint := new(big.Int).Add(proposed, s.basePrice)
	midpoint.Div(midpoint, big.NewInt(2))

	// Never counter below floor.
	if midpoint.Cmp(floor) < 0 {
		return floor
	}
	return midpoint
}

// StrategyMode determines how the agent responds to negotiation proposals.
type StrategyMode string

const (
	StrategyAcceptAll    StrategyMode = "accept_all"
	StrategyRejectAll    StrategyMode = "reject_all"
	StrategyBudgetBound  StrategyMode = "budget_bound"
	StrategyCounterSplit StrategyMode = "counter_split"
)

// StrategyConfig configures the auto-negotiation behavior.
type StrategyConfig struct {
	Strategy StrategyMode `json:"strategy"`
	MaxPrice *big.Int     `json:"maxPrice,omitempty"`
	MinPrice *big.Int     `json:"minPrice,omitempty"`
}

// Decision is the action that AutoNegotiator recommends.
type Decision struct {
	Action ProposalAction
	Terms  Terms
	Reason string
}

// AutoNegotiator evaluates incoming proposals and returns a recommended action.
type AutoNegotiator struct {
	config  StrategyConfig
	pricing PricingQuerier
}

// NewAutoNegotiator creates an auto-negotiator with the given config.
func NewAutoNegotiator(config StrategyConfig, pricing PricingQuerier) *AutoNegotiator {
	return &AutoNegotiator{
		config:  config,
		pricing: pricing,
	}
}

// Evaluate takes a session and the latest incoming proposal, returning a Decision.
func (an *AutoNegotiator) Evaluate(ctx context.Context, session *NegotiationSession, incoming Proposal) (*Decision, error) {
	switch an.config.Strategy {
	case StrategyAcceptAll:
		return an.acceptAll(incoming), nil
	case StrategyRejectAll:
		return an.rejectAll(incoming), nil
	case StrategyBudgetBound:
		return an.budgetBound(ctx, session, incoming)
	case StrategyCounterSplit:
		return an.counterSplit(ctx, session, incoming)
	default:
		return an.rejectAll(incoming), nil
	}
}

func (an *AutoNegotiator) acceptAll(incoming Proposal) *Decision {
	return &Decision{
		Action: ActionAccept,
		Terms:  incoming.Terms,
		Reason: "auto-accept policy",
	}
}

func (an *AutoNegotiator) rejectAll(incoming Proposal) *Decision {
	return &Decision{
		Action: ActionReject,
		Terms:  incoming.Terms,
		Reason: "auto-reject policy",
	}
}

func (an *AutoNegotiator) budgetBound(ctx context.Context, session *NegotiationSession, incoming Proposal) (*Decision, error) {
	maxPrice := an.config.MaxPrice
	if maxPrice == nil && an.pricing != nil {
		quoted, err := an.pricing(incoming.Terms.ToolName, session.InitiatorDID)
		if err != nil {
			return &Decision{
				Action: ActionReject,
				Terms:  incoming.Terms,
				Reason: "price lookup failed",
			}, nil
		}
		maxPrice = quoted
	}

	if maxPrice == nil {
		return &Decision{
			Action: ActionReject,
			Terms:  incoming.Terms,
			Reason: "no max price configured",
		}, nil
	}

	if incoming.Terms.Price.Cmp(maxPrice) <= 0 {
		return &Decision{
			Action: ActionAccept,
			Terms:  incoming.Terms,
			Reason: "within budget",
		}, nil
	}

	return &Decision{
		Action: ActionReject,
		Terms:  incoming.Terms,
		Reason: "exceeds max price",
	}, nil
}

func (an *AutoNegotiator) counterSplit(ctx context.Context, session *NegotiationSession, incoming Proposal) (*Decision, error) {
	maxPrice := an.config.MaxPrice
	if maxPrice == nil && an.pricing != nil {
		quoted, err := an.pricing(incoming.Terms.ToolName, session.InitiatorDID)
		if err != nil {
			return &Decision{
				Action: ActionReject,
				Terms:  incoming.Terms,
				Reason: "price lookup failed",
			}, nil
		}
		maxPrice = quoted
	}

	if maxPrice == nil {
		return &Decision{
			Action: ActionReject,
			Terms:  incoming.Terms,
			Reason: "no max price configured",
		}, nil
	}

	if incoming.Terms.Price.Cmp(maxPrice) <= 0 {
		return &Decision{
			Action: ActionAccept,
			Terms:  incoming.Terms,
			Reason: "within budget",
		}, nil
	}

	if !session.CanCounter() {
		return &Decision{
			Action: ActionReject,
			Terms:  incoming.Terms,
			Reason: "no counter rounds remaining",
		}, nil
	}

	// Split the difference: midpoint between incoming and max.
	midpoint := new(big.Int).Add(incoming.Terms.Price, maxPrice)
	midpoint.Div(midpoint, big.NewInt(2))

	counterTerms := incoming.Terms
	counterTerms.Price = midpoint

	return &Decision{
		Action: ActionCounter,
		Terms:  counterTerms,
		Reason: "counter at midpoint",
	}, nil
}
