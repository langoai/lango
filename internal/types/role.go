package types

// MessageRole represents the role of a message participant.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
	RoleFunction  MessageRole = "function"
	RoleModel     MessageRole = "model"
)

// Valid reports whether r is a known message role.
func (r MessageRole) Valid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleTool, RoleFunction, RoleModel:
		return true
	}
	return false
}

// Values returns all known message roles.
func (r MessageRole) Values() []MessageRole {
	return []MessageRole{RoleUser, RoleAssistant, RoleTool, RoleFunction, RoleModel}
}

// Normalize converts legacy role names to their canonical forms.
// "model" becomes "assistant", "function" becomes "tool".
// All other roles are returned as-is.
func (r MessageRole) Normalize() MessageRole {
	switch r {
	case RoleModel:
		return RoleAssistant
	case RoleFunction:
		return RoleTool
	default:
		return r
	}
}
