package negotiation

import (
	"math/big"
	"testing"
	"time"
)

func TestNegotiatePayload_MarshalRoundTrip(t *testing.T) {
	give := &NegotiatePayload{
		SessionID: "sess-001",
		Proposal: Proposal{
			Action:    ActionPropose,
			SenderDID: "did:lango:buyer123",
			Terms: Terms{
				Price:     big.NewInt(5000000),
				Currency:  "USDC",
				ToolName:  "code-review",
				UseEscrow: true,
			},
			Round:     1,
			Reason:    "initial offer",
			Timestamp: time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC),
		},
	}

	data, err := give.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	got, err := UnmarshalNegotiatePayload(data)
	if err != nil {
		t.Fatalf("UnmarshalNegotiatePayload() error: %v", err)
	}

	if got.SessionID != give.SessionID {
		t.Errorf("SessionID = %q, want %q", got.SessionID, give.SessionID)
	}
	if got.Proposal.Action != give.Proposal.Action {
		t.Errorf("Action = %q, want %q", got.Proposal.Action, give.Proposal.Action)
	}
	if got.Proposal.SenderDID != give.Proposal.SenderDID {
		t.Errorf("SenderDID = %q, want %q", got.Proposal.SenderDID, give.Proposal.SenderDID)
	}
	if got.Proposal.Terms.ToolName != give.Proposal.Terms.ToolName {
		t.Errorf("ToolName = %q, want %q", got.Proposal.Terms.ToolName, give.Proposal.Terms.ToolName)
	}
	if got.Proposal.Terms.Currency != give.Proposal.Terms.Currency {
		t.Errorf("Currency = %q, want %q", got.Proposal.Terms.Currency, give.Proposal.Terms.Currency)
	}
	if got.Proposal.Terms.UseEscrow != give.Proposal.Terms.UseEscrow {
		t.Errorf("UseEscrow = %v, want %v", got.Proposal.Terms.UseEscrow, give.Proposal.Terms.UseEscrow)
	}
	if got.Proposal.Round != give.Proposal.Round {
		t.Errorf("Round = %d, want %d", got.Proposal.Round, give.Proposal.Round)
	}
}

func TestUnmarshalNegotiatePayload_InvalidJSON(t *testing.T) {
	_, err := UnmarshalNegotiatePayload([]byte("not-json"))
	if err == nil {
		t.Error("UnmarshalNegotiatePayload() expected error for invalid JSON")
	}
}

func TestProposalActions(t *testing.T) {
	tests := []struct {
		give ProposalAction
		want string
	}{
		{give: ActionPropose, want: "propose"},
		{give: ActionCounter, want: "counter"},
		{give: ActionAccept, want: "accept"},
		{give: ActionReject, want: "reject"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.give) != tt.want {
				t.Errorf("ProposalAction = %q, want %q", tt.give, tt.want)
			}
		})
	}
}
