package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Observation holds the schema definition for the Observation entity.
// Observation stores raw conversation observations for memory processing.
type Observation struct {
	ent.Schema
}

// Fields of the Observation.
func (Observation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("session_key").
			NotEmpty(),
		field.Text("content").
			NotEmpty(),
		field.Int("token_count").
			Default(0),
		field.Int("source_start_index").
			Default(0),
		field.Int("source_end_index").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Observation.
func (Observation) Edges() []ent.Edge {
	return nil
}

// Indexes of the Observation.
func (Observation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key"),
		index.Fields("created_at"),
	}
}
