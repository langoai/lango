package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// SessionProvenance holds the schema definition for a session tree node.
type SessionProvenance struct {
	ent.Schema
}

// Fields of the SessionProvenance.
func (SessionProvenance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("session_key").
			NotEmpty().
			Unique().
			Comment("Unique session key"),
		field.String("parent_key").
			Optional().
			Comment("Parent session key for tree hierarchy"),
		field.String("agent_name").
			NotEmpty().
			Comment("Name of the agent that owns this session"),
		field.String("goal").
			Optional().
			Comment("Session goal or task description"),
		field.String("run_id").
			Optional().
			Comment("Associated RunLedger run ID"),
		field.String("workspace_id").
			Optional().
			Comment("Associated workspace ID"),
		field.Int("depth").
			Default(0).
			Comment("Depth in the session tree (root = 0)"),
		field.Enum("status").
			Values("active", "merged", "discarded", "completed").
			Default("active").
			Comment("Session lifecycle status"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("closed_at").
			Optional().
			Nillable().
			Comment("When the session was closed (merged/discarded/completed)"),
	}
}

// Edges of the SessionProvenance.
func (SessionProvenance) Edges() []ent.Edge {
	return nil
}

// Indexes of the SessionProvenance.
func (SessionProvenance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_key"),
		index.Fields("agent_name"),
		index.Fields("status"),
		index.Fields("run_id"),
		index.Fields("created_at"),
	}
}
