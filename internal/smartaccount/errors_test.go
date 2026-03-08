package smartaccount

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyViolationError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveID     string
		giveReason string
		wantMsg    string
	}{
		{
			give:       "basic",
			giveID:     "session-123",
			giveReason: "spend limit exceeded",
			wantMsg:    "policy violation for session session-123: spend limit exceeded",
		},
		{
			give:       "empty_id",
			giveID:     "",
			giveReason: "target not allowed",
			wantMsg:    "policy violation for session : target not allowed",
		},
		{
			give:       "empty_reason",
			giveID:     "abc",
			giveReason: "",
			wantMsg:    "policy violation for session abc: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			err := &PolicyViolationError{
				SessionID: tt.giveID,
				Reason:    tt.giveReason,
			}
			assert.Equal(t, tt.wantMsg, err.Error())
		})
	}
}

func TestPolicyViolationError_Unwrap(t *testing.T) {
	t.Parallel()

	err := &PolicyViolationError{
		SessionID: "sess-001",
		Reason:    "function not allowed",
	}

	assert.ErrorIs(t, err, ErrPolicyViolation,
		"Unwrap must return ErrPolicyViolation")
}

func TestPolicyViolationError_ErrorIs(t *testing.T) {
	t.Parallel()

	pve := &PolicyViolationError{
		SessionID: "sess-002",
		Reason:    "exceeded daily limit",
	}

	// errors.Is should match both the concrete error and the sentinel.
	assert.True(t, errors.Is(pve, ErrPolicyViolation))

	// Should not match unrelated sentinels.
	assert.False(t, errors.Is(pve, ErrSessionExpired))
	assert.False(t, errors.Is(pve, ErrSessionRevoked))
	assert.False(t, errors.Is(pve, ErrModuleNotInstalled))
}

func TestPolicyViolationError_ErrorAs(t *testing.T) {
	t.Parallel()

	original := &PolicyViolationError{
		SessionID: "sess-003",
		Reason:    "bad target",
	}

	// Wrap in another error to verify errors.As traversal.
	wrapped := fmt.Errorf("validate call: %w", original)

	var pve *PolicyViolationError
	require.True(t, errors.As(wrapped, &pve),
		"errors.As must find PolicyViolationError through wrapping")
	assert.Equal(t, "sess-003", pve.SessionID)
	assert.Equal(t, "bad target", pve.Reason)
}

func TestPolicyViolationError_DoubleWrap(t *testing.T) {
	t.Parallel()

	original := &PolicyViolationError{
		SessionID: "sess-004",
		Reason:    "rate limited",
	}

	// Double-wrap: PolicyViolationError -> Unwrap -> ErrPolicyViolation
	wrapped := fmt.Errorf("execute: %w",
		fmt.Errorf("check policy: %w", original),
	)

	assert.ErrorIs(t, wrapped, ErrPolicyViolation,
		"must find ErrPolicyViolation through double-wrapped chain")

	var pve *PolicyViolationError
	require.True(t, errors.As(wrapped, &pve))
	assert.Equal(t, "sess-004", pve.SessionID)
}

func TestSentinelErrors_Distinct(t *testing.T) {
	t.Parallel()

	sentinels := []struct {
		give string
		err  error
	}{
		{give: "ErrAccountNotDeployed", err: ErrAccountNotDeployed},
		{give: "ErrSessionExpired", err: ErrSessionExpired},
		{give: "ErrSessionRevoked", err: ErrSessionRevoked},
		{give: "ErrPolicyViolation", err: ErrPolicyViolation},
		{give: "ErrModuleNotInstalled", err: ErrModuleNotInstalled},
		{give: "ErrSpendLimitExceeded", err: ErrSpendLimitExceeded},
		{give: "ErrInvalidSessionKey", err: ErrInvalidSessionKey},
		{give: "ErrSessionNotFound", err: ErrSessionNotFound},
		{give: "ErrTargetNotAllowed", err: ErrTargetNotAllowed},
		{give: "ErrFunctionNotAllowed", err: ErrFunctionNotAllowed},
		{give: "ErrInvalidUserOp", err: ErrInvalidUserOp},
		{give: "ErrBundlerError", err: ErrBundlerError},
		{give: "ErrModuleAlreadyInstalled", err: ErrModuleAlreadyInstalled},
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i == j {
				continue
			}
			assert.NotErrorIs(t, a.err, b.err,
				"%s must not match %s", a.give, b.give)
		}
	}
}

func TestSentinelErrors_NotNil(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		ErrAccountNotDeployed,
		ErrSessionExpired,
		ErrSessionRevoked,
		ErrPolicyViolation,
		ErrModuleNotInstalled,
		ErrSpendLimitExceeded,
		ErrInvalidSessionKey,
		ErrSessionNotFound,
		ErrTargetNotAllowed,
		ErrFunctionNotAllowed,
		ErrInvalidUserOp,
		ErrBundlerError,
		ErrModuleAlreadyInstalled,
	}

	for _, err := range sentinels {
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

func TestSentinelErrors_Wrappable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		giveErr error
	}{
		{give: "ErrAccountNotDeployed", giveErr: ErrAccountNotDeployed},
		{give: "ErrSessionExpired", giveErr: ErrSessionExpired},
		{give: "ErrPolicyViolation", giveErr: ErrPolicyViolation},
		{give: "ErrBundlerError", giveErr: ErrBundlerError},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			wrapped := fmt.Errorf("some context: %w", tt.giveErr)
			assert.ErrorIs(t, wrapped, tt.giveErr,
				"wrapped error must match via errors.Is")
			assert.Contains(t, wrapped.Error(), tt.giveErr.Error())
		})
	}
}

func TestPolicyViolationError_ImplementsError(t *testing.T) {
	t.Parallel()

	var err error = &PolicyViolationError{
		SessionID: "test",
		Reason:    "testing",
	}
	assert.NotEmpty(t, err.Error())
}
