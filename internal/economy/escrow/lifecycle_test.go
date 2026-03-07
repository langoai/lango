package escrow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		give string
		from EscrowStatus
		to   EscrowStatus
		want bool
	}{
		{give: "pending->funded", from: StatusPending, to: StatusFunded, want: true},
		{give: "pending->expired", from: StatusPending, to: StatusExpired, want: true},
		{give: "pending->active (invalid)", from: StatusPending, to: StatusActive, want: false},
		{give: "funded->active", from: StatusFunded, to: StatusActive, want: true},
		{give: "funded->expired", from: StatusFunded, to: StatusExpired, want: true},
		{give: "funded->released (invalid)", from: StatusFunded, to: StatusReleased, want: false},
		{give: "active->completed", from: StatusActive, to: StatusCompleted, want: true},
		{give: "active->disputed", from: StatusActive, to: StatusDisputed, want: true},
		{give: "active->expired", from: StatusActive, to: StatusExpired, want: true},
		{give: "completed->released", from: StatusCompleted, to: StatusReleased, want: true},
		{give: "completed->disputed", from: StatusCompleted, to: StatusDisputed, want: true},
		{give: "disputed->refunded", from: StatusDisputed, to: StatusRefunded, want: true},
		{give: "disputed->released", from: StatusDisputed, to: StatusReleased, want: true},
		{give: "released->anything (terminal)", from: StatusReleased, to: StatusRefunded, want: false},
		{give: "expired->anything (terminal)", from: StatusExpired, to: StatusPending, want: false},
		{give: "refunded->anything (terminal)", from: StatusRefunded, to: StatusPending, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, canTransition(tt.from, tt.to))
		})
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		give    string
		from    EscrowStatus
		to      EscrowStatus
		wantErr bool
	}{
		{give: "valid transition", from: StatusPending, to: StatusFunded, wantErr: false},
		{give: "invalid transition", from: StatusPending, to: StatusReleased, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := validateTransition(tt.from, tt.to)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidTransition)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
