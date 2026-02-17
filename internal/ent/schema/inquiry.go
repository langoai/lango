package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Inquiry holds the schema definition for the Inquiry entity.
// Inquiry stores proactive knowledge questions to ask the user.
type Inquiry struct {
	ent.Schema
}

// Fields of the Inquiry.
func (Inquiry) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("session_key").
			NotEmpty(),
		field.String("topic").
			NotEmpty(),
		field.Text("question").
			NotEmpty(),
		field.Text("context").
			Optional().
			Nillable(),
		field.Enum("priority").
			Values("low", "medium", "high").
			Default("medium"),
		field.Enum("status").
			Values("pending", "resolved", "dismissed").
			Default("pending"),
		field.Text("answer").
			Optional().
			Nillable(),
		field.String("knowledge_key").
			Optional().
			Nillable(),
		field.String("source_observation_id").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("resolved_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Inquiry.
func (Inquiry) Edges() []ent.Edge {
	return nil
}

// Indexes of the Inquiry.
func (Inquiry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key", "status"),
		index.Fields("status"),
	}
}
