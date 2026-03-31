package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ActionLog holds the schema definition for the ActionLog entity.
// ActionLog stores structured execution records for ontology actions.
type ActionLog struct {
	ent.Schema
}

// Fields of the ActionLog.
func (ActionLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("action_name").
			NotEmpty(),
		field.String("principal").
			NotEmpty(),
		field.JSON("params", map[string]string{}).
			Optional(),
		field.Enum("status").
			Values("started", "completed", "failed", "compensated").
			Default("started"),
		field.JSON("effects", map[string]any{}).
			Optional(),
		field.String("error_message").
			Optional().
			Nillable(),
		field.Time("started_at").
			Default(time.Now).
			Immutable(),
		field.Time("completed_at").
			Optional().
			Nillable(),
	}
}

// Edges of the ActionLog.
func (ActionLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the ActionLog.
func (ActionLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("action_name"),
		index.Fields("principal"),
		index.Fields("status"),
		index.Fields("started_at"),
	}
}
