package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CronJob holds the schema definition for a scheduled cron job.
type CronJob struct {
	ent.Schema
}

// Fields of the CronJob.
func (CronJob) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			Unique().
			Comment("Human-readable job name"),
		field.Enum("schedule_type").
			Values("at", "every", "cron").
			Comment("Schedule type: one-time, interval, or cron expression"),
		field.String("schedule").
			NotEmpty().
			Comment("Schedule value: ISO8601 datetime, duration, or cron expression"),
		field.Text("prompt").
			NotEmpty().
			Comment("Prompt to execute when the job fires"),
		field.String("session_mode").
			Default("isolated").
			Comment("Session mode: isolated or main"),
		field.JSON("deliver_to", []string{}).
			Optional().
			Comment("Channels to deliver results to (e.g. slack, telegram)"),
		field.String("timezone").
			Default("UTC").
			Comment("Timezone for schedule evaluation"),
		field.Bool("enabled").
			Default(true),
		field.Int64("timeout_ms").
			Optional().
			Nillable().
			Comment("Per-job timeout in milliseconds (overrides default)"),
		field.Time("last_run_at").
			Optional().
			Nillable().
			Comment("When the job last executed"),
		field.Time("next_run_at").
			Optional().
			Nillable().
			Comment("When the job will next execute"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the CronJob.
func (CronJob) Edges() []ent.Edge {
	return nil
}

// Indexes of the CronJob.
func (CronJob) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("enabled"),
		index.Fields("next_run_at"),
	}
}
