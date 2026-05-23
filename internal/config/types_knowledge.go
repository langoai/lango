package config

import (
	"time"

	"github.com/langoai/lango/internal/types"
)

// KnowledgeConfig defines self-learning knowledge system settings
type KnowledgeConfig struct {
	// Enable the knowledge/learning system
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Maximum context items per layer in retrieval
	MaxContextPerLayer int `mapstructure:"maxContextPerLayer" json:"maxContextPerLayer"`

	// AnalysisTurnThreshold is the number of new turns before triggering conversation analysis (default: 10).
	AnalysisTurnThreshold int `mapstructure:"analysisTurnThreshold" json:"analysisTurnThreshold"`

	// AnalysisTokenThreshold is the token count before triggering conversation analysis (default: 2000).
	AnalysisTokenThreshold int `mapstructure:"analysisTokenThreshold" json:"analysisTokenThreshold"`
}

// ObservationalMemoryConfig defines Observational Memory settings
type ObservationalMemoryConfig struct {
	// Enable the observational memory system
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// LLM provider for observer/reflector (empty = use agent default)
	Provider string `mapstructure:"provider" json:"provider"`

	// Model ID for observer/reflector (empty = use agent default)
	Model string `mapstructure:"model" json:"model"`

	// Token threshold to trigger observation (default: 1000)
	MessageTokenThreshold int `mapstructure:"messageTokenThreshold" json:"messageTokenThreshold"`

	// Token threshold to trigger reflection (default: 2000)
	ObservationTokenThreshold int `mapstructure:"observationTokenThreshold" json:"observationTokenThreshold"`

	// Max token budget for recent messages in context (default: 8000)
	MaxMessageTokenBudget int `mapstructure:"maxMessageTokenBudget" json:"maxMessageTokenBudget"`

	// MaxReflectionsInContext limits reflections injected into LLM context (default: 5, 0 = unlimited).
	MaxReflectionsInContext int `mapstructure:"maxReflectionsInContext" json:"maxReflectionsInContext"`

	// MaxObservationsInContext limits observations injected into LLM context (default: 20, 0 = unlimited).
	MaxObservationsInContext int `mapstructure:"maxObservationsInContext" json:"maxObservationsInContext"`

	// MemoryTokenBudget sets the max token budget for the memory section in system prompt (default: 4000).
	// Zero means use the default.
	MemoryTokenBudget int `mapstructure:"memoryTokenBudget" json:"memoryTokenBudget"`

	// ReflectionConsolidationThreshold is the min reflections before meta-reflection triggers (default: 5).
	// Zero means use the default.
	ReflectionConsolidationThreshold int `mapstructure:"reflectionConsolidationThreshold" json:"reflectionConsolidationThreshold"`
}

// EmbeddingConfig defines embedding and RAG settings.
type EmbeddingConfig struct {
	// Provider selects the embedding provider. Set to "local" for Ollama-based
	// local embeddings, or use a key from the providers map (e.g., "my-openai",
	// "gemini-1") to resolve the backend type and API key automatically.
	Provider string `mapstructure:"provider" json:"provider"`

	// Deprecated: ProviderID is kept only for backwards-compatible config loading.
	// New configs should use Provider for both local and remote providers.
	ProviderID string `mapstructure:"providerID" json:"providerID,omitempty"`

	// Model is the embedding model identifier.
	Model string `mapstructure:"model" json:"model"`

	// Dimensions is the embedding vector dimensionality.
	Dimensions int `mapstructure:"dimensions" json:"dimensions"`

	// Local holds settings for the local (Ollama) provider.
	Local LocalEmbeddingConfig `mapstructure:"local" json:"local"`

	// RAG holds retrieval-augmented generation settings.
	RAG RAGConfig `mapstructure:"rag" json:"rag"`
}

// LocalEmbeddingConfig defines settings for a local embedding provider.
type LocalEmbeddingConfig struct {
	// BaseURL is the Ollama endpoint (default: http://localhost:11434/v1).
	BaseURL string `mapstructure:"baseUrl" json:"baseUrl"`
	// Deprecated: Model is now unified in EmbeddingConfig.Model for all providers.
	// Retained only for backward-compatible config loading and migration.
	Model string `mapstructure:"model" json:"model,omitempty"`
}

// RAGConfig defines retrieval-augmented generation settings.
type RAGConfig struct {
	// Enabled activates RAG context injection.
	Enabled bool `mapstructure:"enabled" json:"enabled"`
	// MaxResults is the maximum number of results to inject.
	MaxResults int `mapstructure:"maxResults" json:"maxResults"`
	// Collections to search (empty means all).
	Collections []string `mapstructure:"collections" json:"collections"`
	// MaxDistance is the maximum cosine distance for RAG results (0.0 = disabled).
	MaxDistance float32 `mapstructure:"maxDistance" json:"maxDistance"`
}

// GraphConfig defines graph store settings for relationship-aware retrieval.
type GraphConfig struct {
	// Enable the graph store.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Backend type: "bolt" (default, embedded BoltDB) or "rocksdb".
	Backend string `mapstructure:"backend" json:"backend"`

	// DatabasePath is the file path for the graph database.
	// Defaults to a "graph.db" file next to the session database.
	DatabasePath string `mapstructure:"databasePath" json:"databasePath"`

	// MaxTraversalDepth limits graph expansion depth (default: 2).
	MaxTraversalDepth int `mapstructure:"maxTraversalDepth" json:"maxTraversalDepth"`

	// MaxExpansionResults limits how many graph-expanded results to return (default: 10).
	MaxExpansionResults int `mapstructure:"maxExpansionResults" json:"maxExpansionResults"`
}

// LibrarianConfig defines proactive knowledge librarian settings.
type LibrarianConfig struct {
	// Enable the proactive librarian system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Minimum observation count to trigger analysis (default: 2).
	ObservationThreshold int `mapstructure:"observationThreshold" json:"observationThreshold"`

	// Turns between inquiries per session (default: 3).
	InquiryCooldownTurns int `mapstructure:"inquiryCooldownTurns" json:"inquiryCooldownTurns"`

	// Maximum pending inquiries per session (default: 2).
	MaxPendingInquiries int `mapstructure:"maxPendingInquiries" json:"maxPendingInquiries"`

	// Minimum confidence level for auto-save: "high", "medium", "low" (default: "high").
	AutoSaveConfidence types.Confidence `mapstructure:"autoSaveConfidence" json:"autoSaveConfidence"`

	// LLM provider for analysis (empty = use agent default).
	Provider string `mapstructure:"provider" json:"provider"`

	// Model ID for analysis (empty = use agent default).
	Model string `mapstructure:"model" json:"model"`
}

// SkillConfig defines file-based skill settings.
type SkillConfig struct {
	// Enable the skill system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// SkillsDir is the directory containing skill files for the knowledge
	// module's internal skill registry (default: ~/.lango/skills).
	//
	// NOTE: Distinct from cfg.Agent.SkillsDir, which feeds the ADK agent's
	// SkillToolset (agentskills.io spec). The two paths are independent and
	// may target the same or different directories per deployment.
	SkillsDir string `mapstructure:"skillsDir" json:"skillsDir"`

	// AllowImport enables importing skills from external URLs and GitHub repositories.
	AllowImport bool `mapstructure:"allowImport" json:"allowImport"`

	// MaxBulkImport limits the number of skills in a single bulk import operation (default: 50).
	MaxBulkImport int `mapstructure:"maxBulkImport" json:"maxBulkImport"`

	// ImportConcurrency sets the number of concurrent HTTP requests during bulk import (default: 5).
	ImportConcurrency int `mapstructure:"importConcurrency" json:"importConcurrency"`

	// ImportTimeout is the overall timeout for skill import operations (default: 2m).
	ImportTimeout time.Duration `mapstructure:"importTimeout" json:"importTimeout"`
}

// ProviderTypeToEmbeddingType maps a provider config type to the corresponding
// embedding backend type.
var ProviderTypeToEmbeddingType = map[types.ProviderType]string{
	types.ProviderOpenAI:    "openai",
	types.ProviderGemini:    "google",
	types.ProviderGoogle:    "google",
	types.ProviderAnthropic: "",
	types.ProviderOllama:    "local",
}

// ResolveEmbeddingProvider returns the embedding backend type and API key
// for the configured embedding provider.
// The Provider field can be "local" (Ollama) or a key in the providers map.
// Legacy configs with ProviderID are handled via MigrateEmbeddingProvider.
func (c *Config) ResolveEmbeddingProvider() (backendType, apiKey string) {
	emb := c.Embedding

	provider := emb.Provider
	// Backwards compatibility: fall back to deprecated ProviderID.
	if provider == "" && emb.ProviderID != "" {
		provider = emb.ProviderID
	}

	if provider == "" {
		return "", ""
	}

	// Local (Ollama) provider — no API key needed.
	if provider == "local" {
		return "local", ""
	}

	// Look up in providers map.
	p, ok := c.Providers[provider]
	if !ok {
		return "", ""
	}
	bt := ProviderTypeToEmbeddingType[p.Type]
	if bt == "" {
		return "", ""
	}
	return bt, p.APIKey
}

// MigrateEmbeddingProvider migrates legacy configs that use separate ProviderID
// and Provider fields into the unified Provider field, and consolidates
// the deprecated Local.Model into the canonical Model field.
func (c *Config) MigrateEmbeddingProvider() {
	if c.Embedding.ProviderID != "" && c.Embedding.Provider == "" {
		c.Embedding.Provider = c.Embedding.ProviderID
		c.Embedding.ProviderID = ""
	}
	// If both are set, Provider takes precedence; clear deprecated ProviderID.
	if c.Embedding.ProviderID != "" && c.Embedding.Provider != "" {
		c.Embedding.ProviderID = ""
	}
	// Migrate deprecated Local.Model into unified Model field.
	if c.Embedding.Local.Model != "" && c.Embedding.Model == "" {
		c.Embedding.Model = c.Embedding.Local.Model
	}
	c.Embedding.Local.Model = ""
}
