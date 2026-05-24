package toolcatalog

import (
	"sort"
	"sync"

	"github.com/langoai/lango/internal/agent"
)

// Category describes a group of related tools.
type Category struct {
	Name        string
	Description string
	ConfigKey   string
	Enabled     bool
}

// ToolEntry pairs a tool with its category.
type ToolEntry struct {
	Tool     *agent.Tool
	Category string
}

// ToolSchema is a summary returned by ListTools (no handler exposed).
type ToolSchema struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Category             string   `json:"category"`
	SafetyLevel          string   `json:"safety_level"`
	Aliases              []string `json:"aliases,omitempty"`
	SearchHints          []string `json:"search_hints,omitempty"`
	Exposure             string   `json:"exposure,omitempty"`
	ReadOnly             bool     `json:"read_only,omitempty"`
	Activity             string   `json:"activity,omitempty"`
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`
}

// Catalog is a thread-safe registry of built-in tools grouped by category.
type Catalog struct {
	mu         sync.RWMutex
	categories map[string]Category
	tools      map[string]ToolEntry
	order      []string
}

// New creates an empty Catalog.
func New() *Catalog {
	return &Catalog{
		categories: make(map[string]Category),
		tools:      make(map[string]ToolEntry),
	}
}

// RegisterCategory adds a category descriptor.
func (c *Catalog) RegisterCategory(cat Category) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.categories[cat.Name] = cat
}

// Register adds tools under the given category.
// The category must already be registered via RegisterCategory.
func (c *Catalog) Register(category string, tools []*agent.Tool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, t := range tools {
		if _, exists := c.tools[t.Name]; !exists {
			c.order = append(c.order, t.Name)
		}
		c.tools[t.Name] = ToolEntry{
			Tool:     t,
			Category: category,
		}
	}
}

// Get returns the entry for the named tool.
func (c *Catalog) Get(name string) (ToolEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.tools[name]
	return e, ok
}

// ListCategories returns all registered categories sorted by name.
func (c *Catalog) ListCategories() []Category {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cats := make([]Category, 0, len(c.categories))
	for _, cat := range c.categories {
		cats = append(cats, cat)
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].Name < cats[j].Name })
	return cats
}

// schemaFromEntry builds a ToolSchema from a ToolEntry, populating capability fields.
func schemaFromEntry(e ToolEntry) ToolSchema {
	cap := e.Tool.Capability
	s := ToolSchema{
		Name:        e.Tool.Name,
		Description: e.Tool.Description,
		Category:    e.Category,
		SafetyLevel: e.Tool.SafetyLevel.String(),
	}
	if len(cap.Aliases) > 0 {
		s.Aliases = cap.Aliases
	}
	if len(cap.SearchHints) > 0 {
		s.SearchHints = cap.SearchHints
	}
	if cap.Exposure != agent.ExposureDefault {
		s.Exposure = cap.Exposure.String()
	}
	s.ReadOnly = cap.ReadOnly
	if cap.Activity != "" {
		s.Activity = string(cap.Activity)
	}
	if len(cap.RequiredCapabilities) > 0 {
		s.RequiredCapabilities = cap.RequiredCapabilities
	}
	return s
}

// SaveableToolNames returns sorted names of tools whose ToolCapability
// indicates they are eligible for knowledge saving (ReadOnly or read/query activity).
func (c *Catalog) SaveableToolNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var names []string
	for _, name := range c.order {
		e := c.tools[name]
		if e.Tool.Capability.KnowledgeSaveable() {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// ListTools returns schemas for all tools in the given category.
// If category is empty, all tools are returned.
func (c *Catalog) ListTools(category string) []ToolSchema {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var schemas []ToolSchema
	for _, name := range c.order {
		e := c.tools[name]
		if category != "" && e.Category != category {
			continue
		}
		schemas = append(schemas, schemaFromEntry(e))
	}
	return schemas
}

// ListVisibleTools returns schemas for tools whose Exposure is visible (Default or AlwaysVisible).
// If category is non-empty, results are further filtered by category.
func (c *Catalog) ListVisibleTools(category string) []ToolSchema {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var schemas []ToolSchema
	for _, name := range c.order {
		e := c.tools[name]
		if category != "" && e.Category != category {
			continue
		}
		if !e.Tool.Capability.Exposure.IsVisible() {
			continue
		}
		schemas = append(schemas, schemaFromEntry(e))
	}
	return schemas
}

// ResolveModeAllowlist expands a mode's tools spec (tool names or "@category"
// references) into a concrete set of tool names. Categories expand to all
// tools in that category. Unknown names are passed through (deferred to
// middleware-level enforcement to surface a clearer error).
func (c *Catalog) ResolveModeAllowlist(modeTools []string) map[string]bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	allow := make(map[string]bool)
	for _, spec := range modeTools {
		if spec == "" {
			continue
		}
		if spec[0] == '@' {
			category := spec[1:]
			for _, name := range c.order {
				if c.tools[name].Category == category {
					allow[name] = true
				}
			}
			continue
		}
		allow[spec] = true
	}
	return allow
}

// ListVisibleToolsForMode returns visible tool schemas filtered by a session
// mode's tool allowlist. modeTools is the raw list from SessionMode.Tools;
// category references ("@foo") are expanded. An empty modeTools slice falls
// back to ListVisibleTools("") (no filtering).
func (c *Catalog) ListVisibleToolsForMode(modeTools []string) []ToolSchema {
	if len(modeTools) == 0 {
		return c.ListVisibleTools("")
	}
	allow := c.ResolveModeAllowlist(modeTools)
	c.mu.RLock()
	defer c.mu.RUnlock()
	var schemas []ToolSchema
	for _, name := range c.order {
		if !allow[name] {
			continue
		}
		e := c.tools[name]
		if !e.Tool.Capability.Exposure.IsVisible() {
			continue
		}
		schemas = append(schemas, schemaFromEntry(e))
	}
	return schemas
}

// SearchableEntries returns all entries except those with ExposureHidden.
// The returned entries are ordered by insertion order.
func (c *Catalog) SearchableEntries() []ToolEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var entries []ToolEntry
	for _, name := range c.order {
		e := c.tools[name]
		if e.Tool.Capability.Exposure == agent.ExposureHidden {
			continue
		}
		entries = append(entries, e)
	}
	return entries
}

// DeferredToolCount returns the number of tools with ExposureDeferred.
func (c *Catalog) DeferredToolCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	for _, e := range c.tools {
		if e.Tool.Capability.Exposure == agent.ExposureDeferred {
			count++
		}
	}
	return count
}

// GetToolSafetyLevel returns the SafetyLevel for the named tool.
// Returns (level, true) if found, or (SafetyLevelDangerous, false) if not found (fail-safe).
func (c *Catalog) GetToolSafetyLevel(name string) (agent.SafetyLevel, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.tools[name]
	if !ok {
		return agent.SafetyLevelDangerous, false
	}
	return e.Tool.SafetyLevel, true
}

// ToolCount returns the total number of registered tools.
func (c *Catalog) ToolCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.tools)
}

// ToolNamesForCategory returns tool names registered under the given category.
func (c *Catalog) ToolNamesForCategory(category string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var names []string
	for _, name := range c.order {
		if c.tools[name].Category == category {
			names = append(names, name)
		}
	}
	return names
}

// EnabledCategorySummary returns a map of enabled category name → tool name list.
func (c *Catalog) EnabledCategorySummary() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	summary := make(map[string][]string)
	for _, cat := range c.categories {
		if !cat.Enabled {
			continue
		}
		var names []string
		for _, n := range c.order {
			if c.tools[n].Category == cat.Name {
				names = append(names, n)
			}
		}
		if len(names) > 0 {
			summary[cat.Name] = names
		}
	}
	return summary
}
