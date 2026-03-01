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
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	SafetyLevel string `json:"safety_level"`
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
		schemas = append(schemas, ToolSchema{
			Name:        e.Tool.Name,
			Description: e.Tool.Description,
			Category:    e.Category,
			SafetyLevel: e.Tool.SafetyLevel.String(),
		})
	}
	return schemas
}

// ToolCount returns the total number of registered tools.
func (c *Catalog) ToolCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.tools)
}
