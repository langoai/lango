package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// RunStep holds the schema definition for a step within a RunLedger run.
// This is a projection for efficient per-step queries.
type RunStep struct {
	ent.Schema
}

// Fields of the RunStep.
func (RunStep) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("run_id").
			NotEmpty().
			Comment("Parent run ID"),
		field.String("step_id").
			NotEmpty().
			Comment("Step ID within the plan"),
		field.Int("step_index").
			Default(0).
			Comment("Display order"),
		field.String("goal").
			Optional(),
		field.String("owner_agent").
			Optional(),
		field.Enum("status").
			Values("pending", "in_progress", "verify_pending", "completed", "failed", "interrupted").
			Default("pending"),
		field.Text("result").
			Optional(),
		field.Text("evidence").
			Optional().
			Comment("JSON-encoded evidence array"),
		field.Text("validator_spec").
			Optional().
			Comment("JSON-encoded ValidatorSpec"),
		field.Int("retry_count").
			Default(0),
		field.Int("max_retries").
			Default(2),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the RunStep.
func (RunStep) Edges() []ent.Edge {
	return nil
}

// Indexes of the RunStep.
func (RunStep) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("run_id"),
		index.Fields("run_id", "step_id").
			Unique(),
		index.Fields("status"),
	}
}
