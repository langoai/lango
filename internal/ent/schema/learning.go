package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Learning holds the schema definition for the Learning entity.
// Learning stores agent learnings from error patterns and fixes.
type Learning struct {
	ent.Schema
}

// Fields of the Learning.
func (Learning) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("trigger").
			NotEmpty(),
		field.Text("error_pattern").
			Optional(),
		field.Text("diagnosis").
			Optional(),
		field.Text("fix").
			Optional(),
		field.Bytes("payload_ciphertext").
			Optional().
			Nillable(),
		field.Bytes("payload_nonce").
			Optional().
			Nillable(),
		field.Int("payload_key_version").
			Optional().
			Nillable(),
		field.Enum("category").
			Values("tool_error", "provider_error", "user_correction", "timeout", "permission", "general"),
		field.JSON("tags", []string{}).
			Optional(),
		field.Int("occurrence_count").
			Default(1),
		field.Int("success_count").
			Default(0),
		field.Float("confidence").
			Default(0.5),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Learning.
func (Learning) Edges() []ent.Edge {
	return nil
}

// Indexes of the Learning.
func (Learning) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category"),
		index.Fields("confidence"),
	}
}
