package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Reflection holds the schema definition for the Reflection entity.
// Reflection stores distilled insights from observations.
type Reflection struct {
	ent.Schema
}

// Fields of the Reflection.
func (Reflection) Fields() []ent.Field {
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
		field.Int("generation").
			Default(1),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Reflection.
func (Reflection) Edges() []ent.Edge {
	return nil
}

// Indexes of the Reflection.
func (Reflection) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key"),
		index.Fields("created_at"),
	}
}
