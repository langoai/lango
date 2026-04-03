package agent

// ToolBuilder provides a fluent API for constructing Tool instances.
type ToolBuilder struct {
	tool Tool
}

// NewTool creates a ToolBuilder with the given name and description.
func NewTool(name, description string) *ToolBuilder {
	return &ToolBuilder{
		tool: Tool{
			Name:        name,
			Description: description,
		},
	}
}

// Safety sets the tool's safety level.
func (b *ToolBuilder) Safety(level SafetyLevel) *ToolBuilder {
	b.tool.SafetyLevel = level
	return b
}

// Params sets the tool's parameter schema.
func (b *ToolBuilder) Params(params map[string]interface{}) *ToolBuilder {
	b.tool.Parameters = params
	return b
}

// Handler sets the tool's execution handler.
func (b *ToolBuilder) Handler(h ToolHandler) *ToolBuilder {
	b.tool.Handler = h
	return b
}

// Aliases sets alternate names for tool search.
func (b *ToolBuilder) Aliases(aliases ...string) *ToolBuilder {
	b.tool.Capability.Aliases = aliases
	return b
}

// Category sets the tool's semantic category.
func (b *ToolBuilder) Category(cat string) *ToolBuilder {
	b.tool.Capability.Category = cat
	return b
}

// Hints sets additional search keywords.
func (b *ToolBuilder) Hints(hints ...string) *ToolBuilder {
	b.tool.Capability.SearchHints = hints
	return b
}

// Deferred sets the tool's exposure policy to ExposureDeferred.
func (b *ToolBuilder) Deferred() *ToolBuilder {
	b.tool.Capability.Exposure = ExposureDeferred
	return b
}

// Hidden sets the tool's exposure policy to ExposureHidden.
func (b *ToolBuilder) Hidden() *ToolBuilder {
	b.tool.Capability.Exposure = ExposureHidden
	return b
}

// ReadOnly marks the tool as performing no mutations.
func (b *ToolBuilder) ReadOnly() *ToolBuilder {
	b.tool.Capability.ReadOnly = true
	return b
}

// ConcurrencySafe marks the tool as safe for concurrent invocation.
func (b *ToolBuilder) ConcurrencySafe() *ToolBuilder {
	b.tool.Capability.ConcurrencySafe = true
	return b
}

// Activity sets the tool's primary activity classification.
func (b *ToolBuilder) Activity(kind ActivityKind) *ToolBuilder {
	b.tool.Capability.Activity = kind
	return b
}

// Requires sets the system capabilities needed by the tool.
func (b *ToolBuilder) Requires(caps ...string) *ToolBuilder {
	b.tool.Capability.RequiredCapabilities = caps
	return b
}

// Build returns the constructed Tool as a pointer.
func (b *ToolBuilder) Build() *Tool {
	t := b.tool
	return &t
}
