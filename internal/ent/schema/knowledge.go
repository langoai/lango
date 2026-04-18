package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Knowledge holds the schema definition for the Knowledge entity.
// Knowledge stores user rules, definitions, preferences, and facts.
type Knowledge struct {
	ent.Schema
}

// Fields of the Knowledge.
func (Knowledge) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("key").
			NotEmpty(),
		field.Enum("category").
			Values("rule", "definition", "preference", "fact", "pattern", "correction"),
		field.Text("content").
			NotEmpty(),
		field.Bytes("content_ciphertext").
			Optional().
			Nillable(),
		field.Bytes("content_nonce").
			Optional().
			Nillable(),
		field.Int("content_key_version").
			Optional().
			Nillable(),
		field.JSON("tags", []string{}).
			Optional(),
		field.String("source").
			Optional(),
		field.Int("version").
			Default(1),
		field.Bool("is_latest").
			Default(true),
		field.Int("use_count").
			Default(0),
		field.Float("relevance_score").
			Default(1.0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Knowledge.
func (Knowledge) Edges() []ent.Edge {
	return nil
}

// Indexes of the Knowledge.
func (Knowledge) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key", "version").Unique(),
		index.Fields("key", "is_latest"),
		index.Fields("category"),
	}
}
