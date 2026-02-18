package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AuditLog holds the schema definition for the AuditLog entity.
// AuditLog stores security audit trail for tool calls, knowledge saves, etc.
type AuditLog struct {
	ent.Schema
}

// Fields of the AuditLog.
func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("session_key").
			Optional(),
		field.Enum("action").
			Values(
				"tool_call",
				"knowledge_save",
				"learning_save",
				"skill_create",
				"skill_execute",
				"skill_import",
				"skill_import_bulk",
				"knowledge_search",
				"approval_request",
				"approval_response",
			),
		field.String("actor").
			NotEmpty(),
		field.String("target").
			Optional(),
		field.JSON("details", map[string]interface{}{}).
			Optional(),
		field.Time("timestamp").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the AuditLog.
func (AuditLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the AuditLog.
func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key"),
		index.Fields("action"),
		index.Fields("timestamp"),
	}
}
