package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// EntityAlias maps a raw entity ID to its canonical ID for entity resolution.
type EntityAlias struct {
	ent.Schema
}

// Fields of the EntityAlias.
func (EntityAlias) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("raw_id").
			NotEmpty().
			Unique(),
		field.String("canonical_id").
			NotEmpty(),
		field.String("source").
			Default("manual"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the EntityAlias.
func (EntityAlias) Edges() []ent.Edge {
	return nil
}

// Indexes of the EntityAlias.
func (EntityAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("canonical_id"),
	}
}
