package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ConfigProfile holds the schema definition for the ConfigProfile entity.
// ConfigProfiles store encrypted application configuration for different environments.
type ConfigProfile struct {
	ent.Schema
}

// Fields of the ConfigProfile.
func (ConfigProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			Unique().
			NotEmpty().
			Comment("Profile identifier (e.g., default, staging, production)"),
		field.Bytes("encrypted_data").
			Comment("AES-256-GCM encrypted JSON configuration blob"),
		field.Bool("active").
			Default(false).
			Comment("Whether this profile is the currently active one"),
		field.Int("version").
			Default(1).
			Comment("Schema version for forward compatibility"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the ConfigProfile.
func (ConfigProfile) Edges() []ent.Edge {
	return nil
}

// Indexes of the ConfigProfile.
func (ConfigProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("active"),
	}
}
