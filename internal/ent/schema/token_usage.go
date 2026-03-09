package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TokenUsage holds the schema definition for the TokenUsage entity.
// TokenUsage stores per-request token usage and estimated cost.
type TokenUsage struct {
	ent.Schema
}

// Fields of the TokenUsage.
func (TokenUsage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("session_key").
			Optional(),
		field.String("provider").
			NotEmpty(),
		field.String("model").
			NotEmpty(),
		field.String("agent_name").
			Optional(),
		field.Int64("input_tokens").
			Default(0),
		field.Int64("output_tokens").
			Default(0),
		field.Int64("total_tokens").
			Default(0),
		field.Int64("cache_tokens").
			Default(0),
		field.Time("timestamp").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the TokenUsage.
func (TokenUsage) Edges() []ent.Edge {
	return nil
}

// Indexes of the TokenUsage.
func (TokenUsage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key"),
		index.Fields("provider"),
		index.Fields("timestamp"),
		index.Fields("agent_name", "timestamp"),
	}
}
