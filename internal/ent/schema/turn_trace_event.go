package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TurnTraceEvent holds the schema definition for an append-only trace event.
type TurnTraceEvent struct {
	ent.Schema
}

// Fields of the TurnTraceEvent.
func (TurnTraceEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("trace_id").
			NotEmpty().
			Immutable(),
		field.Int64("seq").
			Immutable(),
		field.String("event_type").
			NotEmpty(),
		field.String("agent_name").
			Optional(),
		field.String("tool_name").
			Optional(),
		field.String("call_signature").
			Optional(),
		field.Text("payload_json").
			Optional(),
		field.Bool("payload_truncated").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the TurnTraceEvent.
func (TurnTraceEvent) Edges() []ent.Edge {
	return nil
}

// Indexes of the TurnTraceEvent.
func (TurnTraceEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trace_id", "seq").
			Unique(),
		index.Fields("trace_id"),
		index.Fields("event_type"),
		index.Fields("created_at"),
	}
}
