package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ProvenanceCheckpoint holds the schema definition for a provenance checkpoint.
type ProvenanceCheckpoint struct {
	ent.Schema
}

// Fields of the ProvenanceCheckpoint.
func (ProvenanceCheckpoint) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("session_key").
			Optional().
			Comment("Session this checkpoint belongs to"),
		field.String("run_id").
			Optional().
			Comment("RunLedger run ID"),
		field.String("label").
			NotEmpty().
			Comment("Human-readable checkpoint label"),
		field.Enum("trigger").
			Values("manual", "step_complete", "policy_applied").
			Comment("What caused this checkpoint"),
		field.Int64("journal_seq").
			Default(0).
			Comment("RunLedger journal position at checkpoint time"),
		field.String("git_ref").
			Optional().
			Comment("Git commit reference at checkpoint time"),
		field.Text("metadata").
			Optional().
			Comment("JSON-encoded metadata key-value pairs"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the ProvenanceCheckpoint.
func (ProvenanceCheckpoint) Edges() []ent.Edge {
	return nil
}

// Indexes of the ProvenanceCheckpoint.
func (ProvenanceCheckpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_key"),
		index.Fields("run_id"),
		index.Fields("trigger"),
		index.Fields("created_at"),
		index.Fields("run_id", "journal_seq"),
	}
}
