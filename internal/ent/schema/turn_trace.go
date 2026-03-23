package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TurnTrace holds the schema definition for a single user turn trace.
type TurnTrace struct {
	ent.Schema
}

// Fields of the TurnTrace.
func (TurnTrace) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("trace_id").
			NotEmpty().
			Unique().
			Immutable(),
		field.String("session_key").
			NotEmpty(),
		field.String("entrypoint").
			NotEmpty(),
		field.String("outcome").
			Default("running"),
		field.String("error_code").
			Optional(),
		field.Text("summary").
			Optional(),
		field.Time("started_at").
			Default(time.Now).
			Immutable(),
		field.Time("ended_at").
			Optional().
			Nillable(),
	}
}

// Edges of the TurnTrace.
func (TurnTrace) Edges() []ent.Edge {
	return nil
}

// Indexes of the TurnTrace.
func (TurnTrace) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key"),
		index.Fields("entrypoint"),
		index.Fields("outcome"),
		index.Fields("started_at"),
	}
}
