package settings

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewKnowledgeForm creates the Knowledge configuration form.
func NewKnowledgeForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Knowledge Configuration")

	form.AddField(&tuicore.Field{
		Key: "knowledge_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Knowledge.Enabled,
		Description: "Enable the knowledge layer for persistent learning across sessions",
	})

	form.AddField(&tuicore.Field{
		Key: "knowledge_max_context", Label: "Max Context/Layer", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Knowledge.MaxContextPerLayer),
		Description: "Maximum tokens of context injected per knowledge layer per turn",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewSkillForm creates the Skill configuration form.
func NewSkillForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Skill Configuration")

	form.AddField(&tuicore.Field{
		Key: "skill_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Skill.Enabled,
		Description: "Enable file-based skill system for reusable agent capabilities",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_dir", Label: "Skills Directory", Type: tuicore.InputText,
		Value:       cfg.Skill.SkillsDir,
		Placeholder: "~/.lango/skills",
		Description: "Directory where skill YAML files are stored and loaded from",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_allow_import", Label: "Allow Import", Type: tuicore.InputBool,
		Checked:     cfg.Skill.AllowImport,
		Description: "Allow importing skills from external sources (URLs, P2P peers)",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_max_bulk", Label: "Max Bulk Import", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Skill.MaxBulkImport),
		Placeholder: "50",
		Description: "Maximum number of skills to import in a single bulk operation",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "skill_import_concurrency", Label: "Import Concurrency", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Skill.ImportConcurrency),
		Placeholder: "5",
		Description: "Number of skills to import in parallel during bulk operations",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "skill_import_timeout", Label: "Import Timeout", Type: tuicore.InputText,
		Value:       cfg.Skill.ImportTimeout.String(),
		Placeholder: "2m (e.g. 30s, 1m, 5m)",
		Description: "Maximum time allowed for a single skill import operation",
	})

	return &form
}

// NewObservationalMemoryForm creates the Observational Memory configuration form.
func NewObservationalMemoryForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Observational Memory")

	form.AddField(&tuicore.Field{
		Key: "om_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.ObservationalMemory.Enabled,
		Description: "Enable observational memory for automatic user behavior learning",
	})

	omProviderOpts := append([]string{""}, buildProviderOptions(cfg)...)
	omProviderField := &tuicore.Field{
		Key: "om_provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:       cfg.ObservationalMemory.Provider,
		Options:     omProviderOpts,
		Placeholder: "(inherits from Agent)",
		Description: fmt.Sprintf("LLM provider for memory processing; empty = inherit from Agent (%s)", cfg.Agent.Provider),
	}
	form.AddField(omProviderField)

	omModelField := &tuicore.Field{
		Key: "om_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.ObservationalMemory.Model,
		Placeholder: "(inherits from Agent)",
		Description: fmt.Sprintf("Model for observation/reflection generation; empty = inherit from Agent (%s)", cfg.Agent.Model),
	}
	form.AddField(omModelField)

	omFetchProvider := cfg.ObservationalMemory.Provider
	if omFetchProvider == "" {
		omFetchProvider = cfg.Agent.Provider
	}
	if omModelOpts, omErr := FetchModelOptionsWithError(omFetchProvider, cfg, cfg.ObservationalMemory.Model); len(omModelOpts) > 0 {
		omModelOpts = append([]string{""}, omModelOpts...)
		omModelField.Type = tuicore.InputSearchSelect
		omModelField.Options = omModelOpts
		omModelField.Placeholder = ""
	} else if omErr != nil {
		omModelField.Description = fmt.Sprintf("Could not fetch models (%v); enter model ID manually", omErr)
	}

	// Reactive: om_provider → om_model
	omCfgCopy := cfg
	omProviderField.OnChange = func(newProvider string) tea.Cmd {
		fetchP := newProvider
		if fetchP == "" {
			fetchP = omCfgCopy.Agent.Provider
		}
		omModelField.Loading = true
		omModelField.LoadError = nil
		return FetchModelOptionsCmd("om_model", fetchP, omCfgCopy, "")
	}

	form.AddField(&tuicore.Field{
		Key: "om_msg_threshold", Label: "Message Token Threshold",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.MessageTokenThreshold),
		Description: "Minimum tokens in a message before it triggers observation",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_obs_threshold", Label: "Observation Token Threshold",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.ObservationTokenThreshold),
		Description: "Token threshold to trigger consolidation into reflections",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_budget", Label: "Max Message Token Budget",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.MaxMessageTokenBudget),
		Description: "Maximum tokens allocated for memory context in each turn",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_reflections", Label: "Max Reflections in Context",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.MaxReflectionsInContext),
		Description: "Max reflections injected per turn; 0 = unlimited",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer (0 = unlimited)")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_observations", Label: "Max Observations in Context",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.MaxObservationsInContext),
		Description: "Max raw observations injected per turn; 0 = unlimited",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer (0 = unlimited)")
			}
			return nil
		},
	})

	return &form
}

// NewEmbeddingForm creates the Embedding & RAG configuration form.
func NewEmbeddingForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Embedding & RAG Configuration")

	providerOpts := []string{"local"}
	for id := range cfg.Providers {
		providerOpts = append(providerOpts, id)
	}
	sort.Strings(providerOpts)

	embProviderField := &tuicore.Field{
		Key: "emb_provider_id", Label: "Provider", Type: tuicore.InputSelect,
		Value:       cfg.Embedding.Provider,
		Options:     providerOpts,
		Description: "Embedding provider; 'local' uses a local model via Ollama/compatible API",
	}
	form.AddField(embProviderField)

	embModelField := &tuicore.Field{
		Key: "emb_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.Embedding.Model,
		Placeholder: "e.g. text-embedding-3-small",
		Description: "Embedding model name; must be supported by the selected provider",
	}
	form.AddField(embModelField)

	if cfg.Embedding.Provider != "" {
		if embModelOpts := FetchEmbeddingModelOptions(cfg.Embedding.Provider, cfg, cfg.Embedding.Model); len(embModelOpts) > 0 {
			embModelOpts = append([]string{""}, embModelOpts...)
			embModelField.Type = tuicore.InputSearchSelect
			embModelField.Options = embModelOpts
			embModelField.Placeholder = ""
		} else {
			if _, embErr := FetchModelOptionsWithError(cfg.Embedding.Provider, cfg, cfg.Embedding.Model); embErr != nil {
				embModelField.Description = fmt.Sprintf("Could not fetch models (%v); enter model ID manually", embErr)
			}
		}
	}

	// Reactive: emb_provider → emb_model (embedding-filtered)
	embCfgCopy := cfg
	embProviderField.OnChange = func(newProvider string) tea.Cmd {
		if newProvider == "" || newProvider == "local" {
			return nil
		}
		embModelField.Loading = true
		embModelField.LoadError = nil
		return FetchEmbeddingModelOptionsCmd("emb_model", newProvider, embCfgCopy, "")
	}

	form.AddField(&tuicore.Field{
		Key: "emb_dimensions", Label: "Dimensions", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Embedding.Dimensions),
		Description: "Vector dimensions for embeddings; 0 = use model default",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "emb_local_baseurl", Label: "Local Base URL", Type: tuicore.InputText,
		Value:       cfg.Embedding.Local.BaseURL,
		Placeholder: "http://localhost:11434/v1",
		Description: "API base URL for the local embedding server (Ollama, vLLM, etc.)",
	})

	form.AddField(&tuicore.Field{
		Key: "emb_rag_enabled", Label: "RAG Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Embedding.RAG.Enabled,
		Description: "Enable Retrieval-Augmented Generation for knowledge-enhanced responses",
	})

	form.AddField(&tuicore.Field{
		Key: "emb_rag_max_results", Label: "RAG Max Results", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Embedding.RAG.MaxResults),
		Description: "Maximum number of retrieved chunks injected into context per query",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "emb_rag_collections", Label: "RAG Collections", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Embedding.RAG.Collections, ","),
		Placeholder: "collection1,collection2 (comma-separated)",
		Description: "Vector store collections to search during RAG retrieval",
	})

	return &form
}

// NewGraphForm creates the Graph Store configuration form.
func NewGraphForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Graph Store Configuration")

	form.AddField(&tuicore.Field{
		Key: "graph_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Graph.Enabled,
		Description: "Enable knowledge graph for structured entity and relationship storage",
	})

	form.AddField(&tuicore.Field{
		Key: "graph_backend", Label: "Backend", Type: tuicore.InputSelect,
		Value:       cfg.Graph.Backend,
		Options:     []string{"bolt"},
		Description: "Graph database backend; 'bolt' uses embedded BoltDB",
	})

	form.AddField(&tuicore.Field{
		Key: "graph_db_path", Label: "Database Path", Type: tuicore.InputText,
		Value:       cfg.Graph.DatabasePath,
		Placeholder: "~/.lango/graph.db",
		Description: "File path for the graph database storage",
	})

	form.AddField(&tuicore.Field{
		Key: "graph_max_depth", Label: "Max Traversal Depth", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Graph.MaxTraversalDepth),
		Description: "Maximum depth for graph traversal queries (hop count)",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "graph_max_expand", Label: "Max Expansion Results", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Graph.MaxExpansionResults),
		Description: "Maximum nodes returned per expansion query",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewLibrarianForm creates the Librarian configuration form.
func NewLibrarianForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Librarian Configuration")

	form.AddField(&tuicore.Field{
		Key: "lib_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Librarian.Enabled,
		Description: "Enable proactive knowledge extraction from conversations",
	})

	form.AddField(&tuicore.Field{
		Key: "lib_obs_threshold", Label: "Observation Threshold", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Librarian.ObservationThreshold),
		Description: "Minimum observations before the librarian triggers knowledge extraction",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_cooldown", Label: "Inquiry Cooldown Turns", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Librarian.InquiryCooldownTurns),
		Description: "Minimum turns between librarian inquiries to avoid being intrusive",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_max_inquiries", Label: "Max Pending Inquiries", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Librarian.MaxPendingInquiries),
		Description: "Maximum unanswered inquiries before pausing new ones",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_auto_save", Label: "Auto-Save Confidence", Type: tuicore.InputSelect,
		Value:       string(cfg.Librarian.AutoSaveConfidence),
		Options:     []string{"high", "medium", "low"},
		Description: "Confidence threshold for auto-saving extracted knowledge without confirmation",
	})

	libProviderOpts := append([]string{""}, buildProviderOptions(cfg)...)
	libProviderField := &tuicore.Field{
		Key: "lib_provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:       cfg.Librarian.Provider,
		Options:     libProviderOpts,
		Placeholder: "(inherits from Agent)",
		Description: fmt.Sprintf("LLM provider for librarian processing; empty = inherit from Agent (%s)", cfg.Agent.Provider),
	}
	form.AddField(libProviderField)

	libModelField := &tuicore.Field{
		Key: "lib_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.Librarian.Model,
		Placeholder: "(inherits from Agent)",
		Description: fmt.Sprintf("Model for knowledge extraction; empty = inherit from Agent (%s)", cfg.Agent.Model),
	}
	form.AddField(libModelField)

	libFetchProvider := cfg.Librarian.Provider
	if libFetchProvider == "" {
		libFetchProvider = cfg.Agent.Provider
	}
	if libModelOpts, libErr := FetchModelOptionsWithError(libFetchProvider, cfg, cfg.Librarian.Model); len(libModelOpts) > 0 {
		libModelOpts = append([]string{""}, libModelOpts...)
		libModelField.Type = tuicore.InputSearchSelect
		libModelField.Options = libModelOpts
		libModelField.Placeholder = ""
	} else if libErr != nil {
		libModelField.Description = fmt.Sprintf("Could not fetch models (%v); enter model ID manually", libErr)
	}

	// Reactive: lib_provider → lib_model
	libCfgCopy := cfg
	libProviderField.OnChange = func(newProvider string) tea.Cmd {
		fetchP := newProvider
		if fetchP == "" {
			fetchP = libCfgCopy.Agent.Provider
		}
		libModelField.Loading = true
		libModelField.LoadError = nil
		return FetchModelOptionsCmd("lib_model", fetchP, libCfgCopy, "")
	}

	return &form
}
