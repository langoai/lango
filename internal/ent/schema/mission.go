package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Mission holds the schema definition for the durable mission entity.
type Mission struct {
	ent.Schema
}

// Fields of the Mission.
func (Mission) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("session_key").
			NotEmpty(),
		field.String("title").
			NotEmpty(),
		field.Text("description").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("prepared", "active", "waiting_decision", "blocked", "done", "cancelled").
			Default("prepared"),
		field.String("source_kind").
			NotEmpty(),
		field.String("source_ref").
			Optional().
			Nillable(),
		field.String("current_blocked_reason").
			Optional().
			Nillable(),
		field.String("current_decision_kind").
			Optional().
			Nillable(),
		field.Text("current_decision_summary").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("completed_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Mission.
func (Mission) Edges() []ent.Edge {
	return nil
}

// Indexes of the Mission.
func (Mission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key", "updated_at"),
		index.Fields("status"),
		index.Fields("source_kind", "source_ref"),
	}
}
