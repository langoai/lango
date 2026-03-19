package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// RunJournal holds the schema definition for a single journal event
// in the RunLedger append-only log.
type RunJournal struct {
	ent.Schema
}

// Fields of the RunJournal.
func (RunJournal) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("run_id").
			NotEmpty().
			Immutable().
			Comment("Run this event belongs to"),
		field.Int64("seq").
			Immutable().
			Comment("Monotonic sequence number within the run"),
		field.Enum("type").
			Values(
				"run_created",
				"plan_attached",
				"step_started",
				"step_result_proposed",
				"step_validation_passed",
				"step_validation_failed",
				"policy_decision_applied",
				"note_written",
				"run_paused",
				"run_resumed",
				"run_completed",
				"run_failed",
				"projection_synced",
			).
			Comment("Event type"),
		field.Time("timestamp").
			Default(time.Now).
			Immutable(),
		field.Text("payload").
			Comment("JSON-encoded event-specific payload"),
	}
}

// Edges of the RunJournal.
func (RunJournal) Edges() []ent.Edge {
	return nil
}

// Indexes of the RunJournal.
func (RunJournal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("run_id", "seq").
			Unique(),
		index.Fields("run_id"),
		index.Fields("type"),
		index.Fields("timestamp"),
	}
}
