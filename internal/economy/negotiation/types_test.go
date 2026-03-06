package negotiation

import (
	"testing"
)

func TestNegotiationSession_IsTerminal(t *testing.T) {
	tests := []struct {
		give Phase
		want bool
	}{
		{give: PhaseProposed, want: false},
		{give: PhaseCountered, want: false},
		{give: PhaseAccepted, want: true},
		{give: PhaseRejected, want: true},
		{give: PhaseExpired, want: true},
		{give: PhaseCancelled, want: true},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			ns := &NegotiationSession{Phase: tt.give}
			if got := ns.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNegotiationSession_CanCounter(t *testing.T) {
	tests := []struct {
		give     string
		phase    Phase
		round    int
		maxRound int
		want     bool
	}{
		{
			give:     "proposed with rounds remaining",
			phase:    PhaseProposed,
			round:    1,
			maxRound: 3,
			want:     true,
		},
		{
			give:     "countered with rounds remaining",
			phase:    PhaseCountered,
			round:    2,
			maxRound: 3,
			want:     true,
		},
		{
			give:     "max rounds reached",
			phase:    PhaseCountered,
			round:    3,
			maxRound: 3,
			want:     false,
		},
		{
			give:     "terminal phase accepted",
			phase:    PhaseAccepted,
			round:    1,
			maxRound: 3,
			want:     false,
		},
		{
			give:     "terminal phase rejected",
			phase:    PhaseRejected,
			round:    1,
			maxRound: 3,
			want:     false,
		},
		{
			give:     "zero rounds",
			phase:    PhaseProposed,
			round:    0,
			maxRound: 0,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			ns := &NegotiationSession{
				Phase:     tt.phase,
				Round:     tt.round,
				MaxRounds: tt.maxRound,
			}
			if got := ns.CanCounter(); got != tt.want {
				t.Errorf("CanCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}
