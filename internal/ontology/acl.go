package ontology

import (
	"fmt"
	"maps"
	"strings"
)

// ACLPolicy checks whether a principal has the required permission.
type ACLPolicy interface {
	Check(principal string, required Permission) error
}

// AllowAllPolicy permits every operation. Used as the default when ACL is disabled.
type AllowAllPolicy struct{}

func (AllowAllPolicy) Check(string, Permission) error { return nil }

// RoleBasedPolicy maps principal names to their maximum permission level.
// Principals with level >= required pass; others are denied.
// Principals with "peer:" prefix use P2PPermission instead of the roles map.
type RoleBasedPolicy struct {
	roles         map[string]Permission
	p2pPermission Permission // default permission for peer: prefix principals
}

// NewRoleBasedPolicy creates a RoleBasedPolicy from a principal→permission map.
func NewRoleBasedPolicy(roles map[string]Permission) *RoleBasedPolicy {
	m := make(map[string]Permission, len(roles))
	maps.Copy(m, roles)
	return &RoleBasedPolicy{roles: m, p2pPermission: PermRead}
}

// SetP2PPermission sets the default permission for peer: prefix principals.
func (p *RoleBasedPolicy) SetP2PPermission(perm Permission) {
	p.p2pPermission = perm
}

// Check verifies that principal has at least the required permission.
//
// Special rules:
//   - "" or "system" → PermAdmin (programmatic callers without agent context)
//   - Unknown principal (not in roles) → PermRead (safe default)
//   - Known principal → roles[principal] >= required
func (p *RoleBasedPolicy) Check(principal string, required Permission) error {
	if principal == "" || principal == "system" {
		return nil
	}
	// peer: prefix principals use P2PPermission
	if strings.HasPrefix(principal, "peer:") {
		if p.p2pPermission >= required {
			return nil
		}
		return fmt.Errorf("%w: peer principal %q requires %d, has %d", ErrPermissionDenied, principal, required, p.p2pPermission)
	}
	level, ok := p.roles[principal]
	if !ok {
		level = PermRead
	}
	if level >= required {
		return nil
	}
	return fmt.Errorf("%w: principal %q requires %d, has %d", ErrPermissionDenied, principal, required, level)
}

// ParsePermission converts a string ("read", "write", "admin") to a Permission value.
// Returns PermRead for unrecognized strings.
func ParsePermission(s string) Permission {
	switch s {
	case "write":
		return PermWrite
	case "admin":
		return PermAdmin
	default:
		return PermRead
	}
}
