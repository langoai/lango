package agent

// SchemaBuilder provides type-safe construction of JSON Schema objects
// for agent tool parameters. The output is map[string]interface{}
// compatible with agent.Tool.Parameters.
type SchemaBuilder struct {
	props    map[string]interface{}
	required []string
}

// Schema creates a new schema builder for tool parameters.
func Schema() *SchemaBuilder {
	return &SchemaBuilder{
		props: make(map[string]interface{}),
	}
}

// Str adds a string property.
func (b *SchemaBuilder) Str(name, desc string) *SchemaBuilder {
	b.props[name] = map[string]interface{}{
		"type":        "string",
		"description": desc,
	}
	return b
}

// Int adds an integer property.
func (b *SchemaBuilder) Int(name, desc string) *SchemaBuilder {
	b.props[name] = map[string]interface{}{
		"type":        "integer",
		"description": desc,
	}
	return b
}

// Bool adds a boolean property.
func (b *SchemaBuilder) Bool(name, desc string) *SchemaBuilder {
	b.props[name] = map[string]interface{}{
		"type":        "boolean",
		"description": desc,
	}
	return b
}

// Array adds an array property with items of the given type.
func (b *SchemaBuilder) Array(name, itemType, desc string) *SchemaBuilder {
	b.props[name] = map[string]interface{}{
		"type":        "array",
		"description": desc,
		"items":       map[string]interface{}{"type": itemType},
	}
	return b
}

// Enum adds a string property with enumerated values.
func (b *SchemaBuilder) Enum(name, desc string, values ...string) *SchemaBuilder {
	b.props[name] = map[string]interface{}{
		"type":        "string",
		"description": desc,
		"enum":        values,
	}
	return b
}

// Required marks the given field names as required.
func (b *SchemaBuilder) Required(names ...string) *SchemaBuilder {
	b.required = append(b.required, names...)
	return b
}

// Build returns the JSON Schema as map[string]interface{}.
func (b *SchemaBuilder) Build() map[string]interface{} {
	result := map[string]interface{}{
		"type":                 "object",
		"properties":           b.props,
		"additionalProperties": false,
	}
	if len(b.required) > 0 {
		result["required"] = b.required
	}
	return result
}
