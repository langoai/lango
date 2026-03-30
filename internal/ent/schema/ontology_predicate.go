package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OntologyPredicate holds the schema definition for formal relationship types in the ontology.
type OntologyPredicate struct {
	ent.Schema
}

// Fields of the OntologyPredicate.
func (OntologyPredicate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			Unique(),
		field.String("description").
			Default(""),
		field.JSON("source_types", []string{}).
			Optional(),
		field.JSON("target_types", []string{}).
			Optional(),
		field.Enum("cardinality").
			Values("one_to_one", "one_to_many", "many_to_one", "many_to_many").
			Default("many_to_many"),
		field.String("inverse").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("active", "deprecated").
			Default("active"),
		field.Int("version").
			Default(1),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the OntologyPredicate.
func (OntologyPredicate) Edges() []ent.Edge {
	return nil
}

// Indexes of the OntologyPredicate.
func (OntologyPredicate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}
