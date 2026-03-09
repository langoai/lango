package negotiation

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNegotiatePayload_MarshalRoundTrip(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)

	got, err := UnmarshalNegotiatePayload(data)
	require.NoError(t, err)

	assert.Equal(t, give.SessionID, got.SessionID)
	assert.Equal(t, give.Proposal.Action, got.Proposal.Action)
	assert.Equal(t, give.Proposal.SenderDID, got.Proposal.SenderDID)
	assert.Equal(t, give.Proposal.Terms.ToolName, got.Proposal.Terms.ToolName)
	assert.Equal(t, give.Proposal.Terms.Currency, got.Proposal.Terms.Currency)
	assert.Equal(t, give.Proposal.Terms.UseEscrow, got.Proposal.Terms.UseEscrow)
	assert.Equal(t, give.Proposal.Round, got.Proposal.Round)
}

func TestUnmarshalNegotiatePayload_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := UnmarshalNegotiatePayload([]byte("not-json"))
	require.Error(t, err)
}

func TestProposalActions(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			assert.Equal(t, tt.want, string(tt.give))
		})
	}
}
