package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ProvenanceAttribution holds git-aware provenance attribution rows.
type ProvenanceAttribution struct {
	ent.Schema
}

// Fields of the ProvenanceAttribution.
func (ProvenanceAttribution) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("session_key").
			NotEmpty(),
		field.String("run_id").
			Optional(),
		field.String("workspace_id").
			Optional(),
		field.String("author_type").
			NotEmpty(),
		field.String("author_id").
			NotEmpty(),
		field.String("file_path").
			Optional(),
		field.String("commit_hash").
			Optional(),
		field.String("step_id").
			Optional(),
		field.String("source").
			NotEmpty(),
		field.Int("lines_added").
			Default(0),
		field.Int("lines_removed").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the ProvenanceAttribution.
func (ProvenanceAttribution) Edges() []ent.Edge {
	return nil
}

// Indexes of the ProvenanceAttribution.
func (ProvenanceAttribution) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key"),
		index.Fields("workspace_id"),
		index.Fields("run_id"),
		index.Fields("author_id"),
		index.Fields("commit_hash"),
		index.Fields("file_path"),
		index.Fields("created_at"),
	}
}
