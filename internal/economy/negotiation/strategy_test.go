package negotiation

import (
	"context"
	"errors"
	"math/big"
	"testing"
)

func testSession() *NegotiationSession {
	return &NegotiationSession{
		ID:           "s1",
		InitiatorDID: "did:buyer",
		ResponderDID: "did:seller",
		Phase:        PhaseProposed,
		CurrentTerms: &Terms{Price: big.NewInt(5000), Currency: "USDC", ToolName: "code-review"},
		Round:        1,
		MaxRounds:    3,
	}
}

func TestAutoStrategy_GenerateCounter(t *testing.T) {
	tests := []struct {
		give        string
		basePrice   int64
		maxDiscount float64
		proposed    int64
		round       int
		maxRounds   int
		wantPrice   int64
	}{
		{
			give:        "midpoint above floor",
			basePrice:   10000,
			maxDiscount: 0.2,
			proposed:    9000,
			round:       1,
			maxRounds:   3,
			wantPrice:   9500, // (9000+10000)/2 = 9500, floor = 8000
		},
		{
			give:        "midpoint below floor uses floor",
			basePrice:   10000,
			maxDiscount: 0.1, // floor = 9000
			proposed:    2000,
			round:       1,
			maxRounds:   3,
			wantPrice:   9000, // (2000+10000)/2 = 6000, but floor = 9000
		},
		{
			give:        "proposed equals base",
			basePrice:   5000,
			maxDiscount: 0.2,
			proposed:    5000,
			round:       1,
			maxRounds:   3,
			wantPrice:   5000, // (5000+5000)/2 = 5000
		},
		{
			give:        "zero discount",
			basePrice:   5000,
			maxDiscount: 0.0, // floor = 5000
			proposed:    3000,
			round:       1,
			maxRounds:   3,
			wantPrice:   5000, // (3000+5000)/2 = 4000, but floor = 5000
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			s := NewAutoStrategy(big.NewInt(tt.basePrice), tt.maxDiscount)
			got := s.GenerateCounter(big.NewInt(tt.proposed), tt.round, tt.maxRounds)
			if got.Cmp(big.NewInt(tt.wantPrice)) != 0 {
				t.Errorf("GenerateCounter() = %s, want %d", got, tt.wantPrice)
			}
		})
	}
}

func TestAutoNegotiator_AcceptAll(t *testing.T) {
	an := NewAutoNegotiator(StrategyConfig{Strategy: StrategyAcceptAll}, nil)
	ctx := context.Background()

	incoming := Proposal{Terms: Terms{Price: big.NewInt(99999)}}
	d, err := an.Evaluate(ctx, testSession(), incoming)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if d.Action != ActionAccept {
		t.Errorf("Action = %q, want %q", d.Action, ActionAccept)
	}
}

func TestAutoNegotiator_RejectAll(t *testing.T) {
	an := NewAutoNegotiator(StrategyConfig{Strategy: StrategyRejectAll}, nil)
	ctx := context.Background()

	incoming := Proposal{Terms: Terms{Price: big.NewInt(1)}}
	d, err := an.Evaluate(ctx, testSession(), incoming)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if d.Action != ActionReject {
		t.Errorf("Action = %q, want %q", d.Action, ActionReject)
	}
}

func TestAutoNegotiator_BudgetBound(t *testing.T) {
	tests := []struct {
		give       string
		maxPrice   int64
		offerPrice int64
		wantAction ProposalAction
	}{
		{
			give:       "within budget",
			maxPrice:   5000,
			offerPrice: 4000,
			wantAction: ActionAccept,
		},
		{
			give:       "at budget",
			maxPrice:   5000,
			offerPrice: 5000,
			wantAction: ActionAccept,
		},
		{
			give:       "over budget",
			maxPrice:   5000,
			offerPrice: 6000,
			wantAction: ActionReject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			an := NewAutoNegotiator(StrategyConfig{
				Strategy: StrategyBudgetBound,
				MaxPrice: big.NewInt(tt.maxPrice),
			}, nil)

			incoming := Proposal{Terms: Terms{
				Price:    big.NewInt(tt.offerPrice),
				Currency: "USDC",
				ToolName: "code-review",
			}}

			d, err := an.Evaluate(context.Background(), testSession(), incoming)
			if err != nil {
				t.Fatalf("Evaluate() error: %v", err)
			}
			if d.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", d.Action, tt.wantAction)
			}
		})
	}
}

func TestAutoNegotiator_BudgetBound_PricingFallback(t *testing.T) {
	pricing := func(_ string, _ string) (*big.Int, error) {
		return big.NewInt(5000), nil
	}
	an := NewAutoNegotiator(StrategyConfig{Strategy: StrategyBudgetBound}, pricing)

	incoming := Proposal{Terms: Terms{
		Price:    big.NewInt(4000),
		Currency: "USDC",
		ToolName: "code-review",
	}}

	d, err := an.Evaluate(context.Background(), testSession(), incoming)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if d.Action != ActionAccept {
		t.Errorf("Action = %q, want %q", d.Action, ActionAccept)
	}
}

func TestAutoNegotiator_BudgetBound_PricingError(t *testing.T) {
	pricing := func(_ string, _ string) (*big.Int, error) {
		return nil, errors.New("network error")
	}
	an := NewAutoNegotiator(StrategyConfig{Strategy: StrategyBudgetBound}, pricing)

	incoming := Proposal{Terms: Terms{
		Price:    big.NewInt(4000),
		Currency: "USDC",
		ToolName: "code-review",
	}}

	d, err := an.Evaluate(context.Background(), testSession(), incoming)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if d.Action != ActionReject {
		t.Errorf("Action = %q, want %q", d.Action, ActionReject)
	}
	if d.Reason != "price lookup failed" {
		t.Errorf("Reason = %q, want %q", d.Reason, "price lookup failed")
	}
}

func TestAutoNegotiator_BudgetBound_NoMaxPrice(t *testing.T) {
	an := NewAutoNegotiator(StrategyConfig{Strategy: StrategyBudgetBound}, nil)

	incoming := Proposal{Terms: Terms{Price: big.NewInt(100)}}
	d, err := an.Evaluate(context.Background(), testSession(), incoming)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if d.Action != ActionReject {
		t.Errorf("Action = %q, want %q", d.Action, ActionReject)
	}
	if d.Reason != "no max price configured" {
		t.Errorf("Reason = %q, want %q", d.Reason, "no max price configured")
	}
}

func TestAutoNegotiator_CounterSplit(t *testing.T) {
	tests := []struct {
		give       string
		maxPrice   int64
		offerPrice int64
		wantAction ProposalAction
		wantPrice  int64
	}{
		{
			give:       "within budget accepts",
			maxPrice:   6000,
			offerPrice: 5000,
			wantAction: ActionAccept,
			wantPrice:  5000,
		},
		{
			give:       "over budget counters at midpoint",
			maxPrice:   4000,
			offerPrice: 6000,
			wantAction: ActionCounter,
			wantPrice:  5000, // (6000+4000)/2
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			an := NewAutoNegotiator(StrategyConfig{
				Strategy: StrategyCounterSplit,
				MaxPrice: big.NewInt(tt.maxPrice),
			}, nil)

			incoming := Proposal{Terms: Terms{
				Price:    big.NewInt(tt.offerPrice),
				Currency: "USDC",
				ToolName: "code-review",
			}}

			d, err := an.Evaluate(context.Background(), testSession(), incoming)
			if err != nil {
				t.Fatalf("Evaluate() error: %v", err)
			}
			if d.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", d.Action, tt.wantAction)
			}
			if d.Terms.Price.Cmp(big.NewInt(tt.wantPrice)) != 0 {
				t.Errorf("Price = %s, want %d", d.Terms.Price, tt.wantPrice)
			}
		})
	}
}

func TestAutoNegotiator_CounterSplit_NoRoundsLeft(t *testing.T) {
	an := NewAutoNegotiator(StrategyConfig{
		Strategy: StrategyCounterSplit,
		MaxPrice: big.NewInt(3000),
	}, nil)

	session := testSession()
	session.Round = 3
	session.MaxRounds = 3

	incoming := Proposal{Terms: Terms{
		Price:    big.NewInt(5000),
		Currency: "USDC",
		ToolName: "code-review",
	}}

	d, err := an.Evaluate(context.Background(), session, incoming)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if d.Action != ActionReject {
		t.Errorf("Action = %q, want %q", d.Action, ActionReject)
	}
	if d.Reason != "no counter rounds remaining" {
		t.Errorf("Reason = %q, want %q", d.Reason, "no counter rounds remaining")
	}
}

func TestAutoNegotiator_UnknownStrategy(t *testing.T) {
	an := NewAutoNegotiator(StrategyConfig{Strategy: "unknown"}, nil)

	incoming := Proposal{Terms: Terms{Price: big.NewInt(100)}}
	d, err := an.Evaluate(context.Background(), testSession(), incoming)
	if err != nil {
		t.Fatalf("Evaluate() error: %v", err)
	}
	if d.Action != ActionReject {
		t.Errorf("Action = %q, want %q", d.Action, ActionReject)
	}
}
