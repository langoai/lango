package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ToolCall represents a tool invocation (embedded in Message)
type ToolCall struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output,omitempty"`
}

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.String("role").
			NotEmpty(),
		field.Text("content"),
		field.Time("timestamp").
			Default(time.Now),
		field.JSON("tool_calls", []ToolCall{}).
			Optional(),
		field.String("author").
			Optional().
			Default(""),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", Session.Type).
			Ref("messages").
			Unique(),
	}
}
