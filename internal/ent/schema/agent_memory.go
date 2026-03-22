package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AgentMemory holds the schema definition for the AgentMemory entity.
// AgentMemory stores per-agent persistent memory entries such as patterns,
// preferences, facts, and skills.
type AgentMemory struct {
	ent.Schema
}

// Fields of the AgentMemory.
func (AgentMemory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("agent_name").
			NotEmpty(),
		field.Enum("scope").
			Values("instance", "global"),
		field.Enum("kind").
			Values("pattern", "preference", "fact", "skill"),
		field.String("key").
			NotEmpty(),
		field.Text("content").
			NotEmpty(),
		field.Float("confidence").
			Default(0.5),
		field.Int("use_count").
			Default(0),
		field.JSON("tags", []string{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the AgentMemory.
func (AgentMemory) Edges() []ent.Edge {
	return nil
}

// Indexes of the AgentMemory.
func (AgentMemory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_name", "key").
			Unique(),
		index.Fields("agent_name"),
		index.Fields("scope"),
		index.Fields("kind"),
		index.Fields("confidence"),
	}
}
