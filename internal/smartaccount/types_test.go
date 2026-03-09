package smartaccount

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModuleType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give ModuleType
		want string
	}{
		{give: ModuleTypeValidator, want: "validator"},
		{give: ModuleTypeExecutor, want: "executor"},
		{give: ModuleTypeFallback, want: "fallback"},
		{give: ModuleTypeHook, want: "hook"},
		{give: ModuleType(0), want: "unknown"},
		{give: ModuleType(255), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestSessionKey_IsMaster(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want bool
	}{
		{give: "", want: true},
		{give: "parent-123", want: false},
		{give: "any-non-empty", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			sk := &SessionKey{ParentID: tt.give}
			assert.Equal(t, tt.want, sk.IsMaster())
		})
	}
}

func TestSessionKey_IsExpired(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		give string
		exp  time.Time
		want bool
	}{
		{give: "expired_1h_ago", exp: now.Add(-time.Hour), want: true},
		{give: "expires_1h_later", exp: now.Add(time.Hour), want: false},
		{give: "expired_1s_ago", exp: now.Add(-time.Second), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			sk := &SessionKey{ExpiresAt: tt.exp}
			assert.Equal(t, tt.want, sk.IsExpired())
		})
	}
}

func TestSessionKey_IsActive(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	tests := []struct {
		give    string
		revoked bool
		exp     time.Time
		want    bool
	}{
		{give: "active_not_revoked_not_expired", revoked: false, exp: future, want: true},
		{give: "revoked_not_expired", revoked: true, exp: future, want: false},
		{give: "not_revoked_expired", revoked: false, exp: past, want: false},
		{give: "revoked_and_expired", revoked: true, exp: past, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			sk := &SessionKey{
				Revoked:   tt.revoked,
				ExpiresAt: tt.exp,
			}
			assert.Equal(t, tt.want, sk.IsActive())
		})
	}
}
