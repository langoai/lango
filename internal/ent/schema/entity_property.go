package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// EntityProperty stores per-entity property values in an EAV model.
type EntityProperty struct {
	ent.Schema
}

// Fields of the EntityProperty.
func (EntityProperty) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("entity_id").
			NotEmpty(),
		field.String("entity_type").
			NotEmpty(),
		field.String("property").
			NotEmpty(),
		field.Text("value"),
		field.String("value_type").
			Default("string"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the EntityProperty.
func (EntityProperty) Edges() []ent.Edge {
	return nil
}

// Indexes of the EntityProperty.
func (EntityProperty) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("entity_id", "property").Unique(),
		index.Fields("entity_type", "property", "value"),
		index.Fields("entity_type"),
	}
}
