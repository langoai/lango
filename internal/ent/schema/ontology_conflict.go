package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OntologyConflict holds the schema for detected contradictions between triples.
type OntologyConflict struct {
	ent.Schema
}

// Fields of the OntologyConflict.
func (OntologyConflict) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("subject").
			NotEmpty(),
		field.String("predicate").
			NotEmpty(),
		field.JSON("candidates", []map[string]interface{}{}).
			Optional().
			Comment("JSON array of CandidateTriple snapshots"),
		field.Enum("status").
			Values("open", "resolved", "auto_resolved").
			Default("open"),
		field.String("resolution").
			Optional().
			Nillable(),
		field.Time("resolved_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the OntologyConflict.
func (OntologyConflict) Edges() []ent.Edge {
	return nil
}

// Indexes of the OntologyConflict.
func (OntologyConflict) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subject", "predicate"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
