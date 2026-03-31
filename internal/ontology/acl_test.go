package ontology

import (
	"context"
	"errors"
	"testing"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowAllPolicy(t *testing.T) {
	p := AllowAllPolicy{}
	tests := []struct {
		give     string
		wantPerm Permission
	}{
		{"ontologist", PermAdmin},
		{"chronicler", PermRead},
		{"", PermAdmin},
		{"unknown", PermWrite},
	}
	for _, tt := range tests {
		assert.NoError(t, p.Check(tt.give, tt.wantPerm))
	}
}

func TestRoleBasedPolicy_Grant(t *testing.T) {
	p := NewRoleBasedPolicy(map[string]Permission{
		"ontologist": PermAdmin,
		"operator":   PermWrite,
		"chronicler": PermRead,
	})

	tests := []struct {
		give     string
		givePerm Permission
		wantErr  bool
	}{
		// admin can do everything
		{"ontologist", PermRead, false},
		{"ontologist", PermWrite, false},
		{"ontologist", PermAdmin, false},
		// write can read and write
		{"operator", PermRead, false},
		{"operator", PermWrite, false},
		{"operator", PermAdmin, true},
		// read can only read
		{"chronicler", PermRead, false},
		{"chronicler", PermWrite, true},
		{"chronicler", PermAdmin, true},
	}
	for _, tt := range tests {
		t.Run(tt.give+"_"+permName(tt.givePerm), func(t *testing.T) {
			err := p.Check(tt.give, tt.givePerm)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrPermissionDenied))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRoleBasedPolicy_SystemFull(t *testing.T) {
	p := NewRoleBasedPolicy(map[string]Permission{
		"chronicler": PermRead,
	})

	assert.NoError(t, p.Check("", PermAdmin), "empty string = system = full access")
	assert.NoError(t, p.Check("system", PermAdmin), "system principal = full access")
}

func TestRoleBasedPolicy_UnknownReadOnly(t *testing.T) {
	p := NewRoleBasedPolicy(map[string]Permission{
		"ontologist": PermAdmin,
	})

	assert.NoError(t, p.Check("unknown_agent", PermRead))
	assert.Error(t, p.Check("unknown_agent", PermWrite))
	assert.Error(t, p.Check("unknown_agent", PermAdmin))
}

func TestParsePermission(t *testing.T) {
	tests := []struct {
		give string
		want Permission
	}{
		{"read", PermRead},
		{"write", PermWrite},
		{"admin", PermAdmin},
		{"unknown", PermRead},
		{"", PermRead},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, ParsePermission(tt.give))
	}
}

// --- Service-level ACL integration tests ---
// These are in the ontology package (not ontology_test) to access ServiceImpl directly.

func TestServiceImpl_checkPermission_NilACL(t *testing.T) {
	svc := &ServiceImpl{}
	ctx := context.Background()
	assert.NoError(t, svc.checkPermission(ctx, PermAdmin), "nil acl = allow all")
}

func TestServiceImpl_checkPermission_WithACL(t *testing.T) {
	svc := &ServiceImpl{}
	svc.SetACLPolicy(NewRoleBasedPolicy(map[string]Permission{
		"chronicler": PermRead,
		"ontologist": PermAdmin,
	}))

	tests := []struct {
		give      string
		givePerm  Permission
		wantError bool
	}{
		// system (empty ctx) = full access
		{"", PermAdmin, false},
		// chronicler = read only
		{"chronicler", PermRead, false},
		{"chronicler", PermWrite, true},
		{"chronicler", PermAdmin, true},
		// ontologist = admin
		{"ontologist", PermRead, false},
		{"ontologist", PermWrite, false},
		{"ontologist", PermAdmin, false},
	}
	for _, tt := range tests {
		t.Run(tt.give+"_"+permName(tt.givePerm), func(t *testing.T) {
			ctx := context.Background()
			if tt.give != "" {
				ctx = ctxkeys.WithPrincipal(ctx, tt.give)
			}
			err := svc.checkPermission(ctx, tt.givePerm)
			if tt.wantError {
				assert.ErrorIs(t, err, ErrPermissionDenied)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func permName(p Permission) string {
	switch p {
	case PermRead:
		return "read"
	case PermWrite:
		return "write"
	case PermAdmin:
		return "admin"
	default:
		return "unknown"
	}
}
