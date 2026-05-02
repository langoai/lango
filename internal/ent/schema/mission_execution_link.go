package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MissionExecutionLink holds the schema definition for mission-to-execution links.
type MissionExecutionLink struct {
	ent.Schema
}

// Fields of the MissionExecutionLink.
func (MissionExecutionLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("mission_id", uuid.UUID{}),
		field.Enum("execution_kind").
			Values("runledger_run", "task_os_execution"),
		field.String("execution_ref").
			NotEmpty(),
		field.Enum("link_role").
			Values("primary", "followup", "retry", "research", "draft", "handoff").
			Default("primary"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MissionExecutionLink.
func (MissionExecutionLink) Edges() []ent.Edge {
	return nil
}

// Indexes of the MissionExecutionLink.
func (MissionExecutionLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mission_id", "execution_kind", "execution_ref").
			Unique(),
		index.Fields("execution_kind", "execution_ref"),
		index.Fields("mission_id", "link_role"),
	}
}
