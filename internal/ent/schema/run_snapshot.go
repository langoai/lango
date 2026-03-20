package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// RunSnapshot holds the schema definition for a cached RunLedger snapshot.
// This is a materialized view — the journal is the source of truth.
type RunSnapshot struct {
	ent.Schema
}

// Fields of the RunSnapshot.
func (RunSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("run_id").
			NotEmpty().
			Unique().
			Comment("Run this snapshot belongs to"),
		field.String("session_key").
			Optional().
			Comment("Session that owns this run"),
		field.Enum("status").
			Values("planning", "running", "paused", "completed", "failed").
			Default("planning"),
		field.String("goal").
			Optional(),
		field.Text("snapshot_data").
			Comment("Full JSON-serialized RunSnapshot"),
		field.Int64("last_journal_seq").
			Default(0).
			Comment("Last journal event seq applied to this snapshot"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the RunSnapshot.
func (RunSnapshot) Edges() []ent.Edge {
	return nil
}

// Indexes of the RunSnapshot.
func (RunSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("session_key"),
		index.Fields("updated_at"),
	}
}
