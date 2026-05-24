package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MissionStateHistory holds the schema definition for append-only mission state transitions.
type MissionStateHistory struct {
	ent.Schema
}

// Fields of the MissionStateHistory.
func (MissionStateHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("mission_id", uuid.UUID{}),
		field.Int64("seq"),
		field.Enum("from_status").
			Values("prepared", "active", "waiting_decision", "blocked", "done", "cancelled").
			Optional().
			Nillable(),
		field.Enum("to_status").
			Values("prepared", "active", "waiting_decision", "blocked", "done", "cancelled"),
		field.String("reason").
			Optional().
			Nillable(),
		field.String("actor_kind").
			NotEmpty(),
		field.String("actor_ref").
			Optional().
			Nillable(),
		field.String("execution_kind").
			Optional().
			Nillable(),
		field.String("execution_ref").
			Optional().
			Nillable(),
		field.String("decision_kind").
			Optional().
			Nillable(),
		field.Text("decision_summary").
			Optional().
			Nillable(),
		field.JSON("payload", map[string]any{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MissionStateHistory.
func (MissionStateHistory) Edges() []ent.Edge {
	return nil
}

// Indexes of the MissionStateHistory.
func (MissionStateHistory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mission_id", "seq").
			Unique(),
		index.Fields("mission_id", "created_at"),
	}
}
