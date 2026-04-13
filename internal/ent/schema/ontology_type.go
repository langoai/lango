package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OntologyType holds the schema definition for formal entity types in the ontology.
type OntologyType struct {
	ent.Schema
}

// Fields of the OntologyType.
func (OntologyType) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			Unique(),
		field.String("description").
			Default(""),
		field.JSON("properties", []map[string]interface{}{}).
			Optional(),
		field.String("extends").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("proposed", "quarantined", "shadow", "active", "deprecated").
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

// Edges of the OntologyType.
func (OntologyType) Edges() []ent.Edge {
	return nil
}

// Indexes of the OntologyType.
func (OntologyType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}
