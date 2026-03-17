package settings

import (
	"github.com/langoai/lango/internal/config"
)

// DepStatus represents the status of a dependency.
type DepStatus int

const (
	// DepMet means the dependency is satisfied.
	DepMet DepStatus = iota
	// DepNotEnabled means the required feature is not enabled.
	DepNotEnabled
	// DepMisconfigured means the feature is enabled but misconfigured.
	DepMisconfigured
)

// Dependency describes a single prerequisite for a feature category.
type Dependency struct {
	// CategoryID is the settings category that must be configured.
	CategoryID string
	// Label is a human-readable name shown in the prerequisite panel.
	Label string
	// Required marks whether the dependency is mandatory (vs. optional/enhancing).
	Required bool
	// Check evaluates the dependency against the current config.
	Check func(cfg *config.Config) DepStatus
	// FixHint is a short hint shown when the dependency is unmet.
	FixHint string
}

// DepResult holds the evaluated result of a single dependency check.
type DepResult struct {
	Dependency
	Status DepStatus
}

// DependencyIndex provides O(1) lookup of dependencies by target category ID.
type DependencyIndex struct {
	// deps maps target category ID → its list of dependencies.
	deps map[string][]Dependency
	// reverse maps dependency category ID → list of category IDs that depend on it.
	reverse map[string][]string
}

// NewDependencyIndex builds the index from the default dependency graph.
func NewDependencyIndex() *DependencyIndex {
	idx := &DependencyIndex{
		deps:    make(map[string][]Dependency),
		reverse: make(map[string][]string),
	}
	for target, deps := range defaultDependencies() {
		idx.deps[target] = deps
		for _, d := range deps {
			idx.reverse[d.CategoryID] = append(idx.reverse[d.CategoryID], target)
		}
	}
	return idx
}

// Evaluate returns evaluated dependency results for the given category.
func (idx *DependencyIndex) Evaluate(categoryID string, cfg *config.Config) []DepResult {
	deps := idx.deps[categoryID]
	if len(deps) == 0 {
		return nil
	}
	results := make([]DepResult, len(deps))
	for i, d := range deps {
		results[i] = DepResult{
			Dependency: d,
			Status:     d.Check(cfg),
		}
	}
	return results
}

// UnmetRequired returns the count of unmet required dependencies for a category.
func (idx *DependencyIndex) UnmetRequired(categoryID string, cfg *config.Config) int {
	count := 0
	for _, d := range idx.deps[categoryID] {
		if d.Required && d.Check(cfg) != DepMet {
			count++
		}
	}
	return count
}

// AllTransitiveUnmet collects all transitively unmet required dependencies.
// It guards against cycles via a visited set.
func (idx *DependencyIndex) AllTransitiveUnmet(categoryID string, cfg *config.Config) []DepResult {
	visited := make(map[string]bool)
	return idx.collectTransitive(categoryID, cfg, visited)
}

func (idx *DependencyIndex) collectTransitive(categoryID string, cfg *config.Config, visited map[string]bool) []DepResult {
	if visited[categoryID] {
		return nil
	}
	visited[categoryID] = true

	var result []DepResult
	for _, d := range idx.deps[categoryID] {
		status := d.Check(cfg)
		if !d.Required || status == DepMet {
			continue
		}
		// First collect children (depth-first) so they appear before the parent.
		children := idx.collectTransitive(d.CategoryID, cfg, visited)
		result = append(result, children...)
		result = append(result, DepResult{Dependency: d, Status: status})
	}
	return result
}

// Dependents returns the list of category IDs that depend on the given category.
func (idx *DependencyIndex) Dependents(categoryID string) []string {
	return idx.reverse[categoryID]
}

// HasDependencies returns true if the category has any registered dependencies.
func (idx *DependencyIndex) HasDependencies(categoryID string) bool {
	return len(idx.deps[categoryID]) > 0
}

// defaultDependencies defines all known feature dependency relationships.
func defaultDependencies() map[string][]Dependency {
	return map[string][]Dependency{
		// Smart Account depends on Payment + Security Signer
		"smartaccount": {
			{
				CategoryID: "payment",
				Label:      "Payment",
				Required:   true,
				FixHint:    "Enable Payment and configure wallet provider",
				Check: func(cfg *config.Config) DepStatus {
					if !cfg.Payment.Enabled {
						return DepNotEnabled
					}
					if cfg.Payment.Network.RPCURL == "" {
						return DepMisconfigured
					}
					return DepMet
				},
			},
			{
				CategoryID: "security",
				Label:      "Security Signer",
				Required:   true,
				FixHint:    "Configure a security signer provider (local/rpc)",
				Check: func(cfg *config.Config) DepStatus {
					if cfg.Security.Signer.Provider == "" {
						return DepNotEnabled
					}
					return DepMet
				},
			},
			{
				CategoryID: "economy",
				Label:      "Economy",
				Required:   false,
				FixHint:    "Enable Economy for budget management",
				Check: func(cfg *config.Config) DepStatus {
					if !cfg.Economy.Enabled {
						return DepNotEnabled
					}
					return DepMet
				},
			},
		},

		// SA sub-categories depend on Smart Account
		"smartaccount_session": {
			{
				CategoryID: "smartaccount",
				Label:      "Smart Account",
				Required:   true,
				FixHint:    "Enable Smart Account first",
				Check:      checkSmartAccountEnabled,
			},
		},
		"smartaccount_paymaster": {
			{
				CategoryID: "smartaccount",
				Label:      "Smart Account",
				Required:   true,
				FixHint:    "Enable Smart Account first",
				Check:      checkSmartAccountEnabled,
			},
		},
		"smartaccount_modules": {
			{
				CategoryID: "smartaccount",
				Label:      "Smart Account",
				Required:   true,
				FixHint:    "Enable Smart Account first",
				Check:      checkSmartAccountEnabled,
			},
		},

		// P2P Network depends on Security Signer (wallet)
		"p2p": {
			{
				CategoryID: "security",
				Label:      "Security Signer",
				Required:   true,
				FixHint:    "Configure a security signer for node identity",
				Check: func(cfg *config.Config) DepStatus {
					if cfg.Security.Signer.Provider == "" {
						return DepNotEnabled
					}
					return DepMet
				},
			},
		},

		// P2P sub-categories depend on P2P
		"p2p_workspace": {
			{
				CategoryID: "p2p",
				Label:      "P2P Network",
				Required:   true,
				FixHint:    "Enable P2P networking first",
				Check:      checkP2PEnabled,
			},
		},
		"p2p_zkp": {
			{
				CategoryID: "p2p",
				Label:      "P2P Network",
				Required:   true,
				FixHint:    "Enable P2P networking first",
				Check:      checkP2PEnabled,
			},
		},
		"p2p_sandbox": {
			{
				CategoryID: "p2p",
				Label:      "P2P Network",
				Required:   true,
				FixHint:    "Enable P2P networking first",
				Check:      checkP2PEnabled,
			},
		},
		"p2p_owner": {
			{
				CategoryID: "p2p",
				Label:      "P2P Network",
				Required:   true,
				FixHint:    "Enable P2P networking first",
				Check:      checkP2PEnabled,
			},
		},

		// P2P Pricing depends on P2P + Payment
		"p2p_pricing": {
			{
				CategoryID: "p2p",
				Label:      "P2P Network",
				Required:   true,
				FixHint:    "Enable P2P networking first",
				Check:      checkP2PEnabled,
			},
			{
				CategoryID: "payment",
				Label:      "Payment",
				Required:   true,
				FixHint:    "Enable Payment for paid tool invocations",
				Check: func(cfg *config.Config) DepStatus {
					if !cfg.Payment.Enabled {
						return DepNotEnabled
					}
					return DepMet
				},
			},
		},

		// Librarian depends on Knowledge + Observational Memory
		"librarian": {
			{
				CategoryID: "knowledge",
				Label:      "Knowledge",
				Required:   true,
				FixHint:    "Enable Knowledge system for extraction",
				Check: func(cfg *config.Config) DepStatus {
					if !cfg.Knowledge.Enabled {
						return DepNotEnabled
					}
					return DepMet
				},
			},
			{
				CategoryID: "observational_memory",
				Label:      "Observational Memory",
				Required:   true,
				FixHint:    "Enable Observational Memory for observation tracking",
				Check: func(cfg *config.Config) DepStatus {
					if !cfg.ObservationalMemory.Enabled {
						return DepNotEnabled
					}
					return DepMet
				},
			},
		},

		// Economy sub-categories depend on Economy
		"economy_risk": {
			{
				CategoryID: "economy",
				Label:      "Economy",
				Required:   true,
				FixHint:    "Enable Economy layer first",
				Check:      checkEconomyEnabled,
			},
		},
		"economy_negotiation": {
			{
				CategoryID: "economy",
				Label:      "Economy",
				Required:   true,
				FixHint:    "Enable Economy layer first",
				Check:      checkEconomyEnabled,
			},
		},
		"economy_pricing": {
			{
				CategoryID: "economy",
				Label:      "Economy",
				Required:   true,
				FixHint:    "Enable Economy layer first",
				Check:      checkEconomyEnabled,
			},
		},
		"economy_escrow": {
			{
				CategoryID: "economy",
				Label:      "Economy",
				Required:   true,
				FixHint:    "Enable Economy layer first",
				Check:      checkEconomyEnabled,
			},
		},

		// Economy Escrow On-Chain depends on Economy + Payment
		"economy_escrow_onchain": {
			{
				CategoryID: "economy",
				Label:      "Economy",
				Required:   true,
				FixHint:    "Enable Economy layer first",
				Check:      checkEconomyEnabled,
			},
			{
				CategoryID: "payment",
				Label:      "Payment",
				Required:   true,
				FixHint:    "Enable Payment for on-chain settlement",
				Check: func(cfg *config.Config) DepStatus {
					if !cfg.Payment.Enabled {
						return DepNotEnabled
					}
					return DepMet
				},
			},
		},

		// Embedding/RAG requires embedding provider
		"embedding": {
			{
				CategoryID: "embedding",
				Label:      "Embedding Provider",
				Required:   true,
				FixHint:    "Set an embedding provider (e.g., local, openai)",
				Check: func(cfg *config.Config) DepStatus {
					// Self-check: this is always met since the form sets it.
					// This entry exists for documentation; actual check is no-op.
					return DepMet
				},
			},
		},

		// Graph depends on embedding
		"graph": {
			{
				CategoryID: "embedding",
				Label:      "Embedding & RAG",
				Required:   false,
				FixHint:    "Configure embedding provider for GraphRAG support",
				Check: func(cfg *config.Config) DepStatus {
					if cfg.Embedding.Provider == "" {
						return DepNotEnabled
					}
					return DepMet
				},
			},
		},
	}
}

// Shared check functions to avoid duplication.

func checkSmartAccountEnabled(cfg *config.Config) DepStatus {
	if !cfg.SmartAccount.Enabled {
		return DepNotEnabled
	}
	return DepMet
}

func checkP2PEnabled(cfg *config.Config) DepStatus {
	if !cfg.P2P.Enabled {
		return DepNotEnabled
	}
	return DepMet
}

func checkEconomyEnabled(cfg *config.Config) DepStatus {
	if !cfg.Economy.Enabled {
		return DepNotEnabled
	}
	return DepMet
}
