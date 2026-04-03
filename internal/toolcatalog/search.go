package toolcatalog

import (
	"sort"
	"strings"

	"github.com/langoai/lango/internal/agent"
)

// SearchIndex provides weighted keyword search over the tool catalog.
type SearchIndex struct {
	entries []searchEntry
}

type searchEntry struct {
	Name        string
	Aliases     []string
	Description string
	CatalogCat  string // catalog-level category
	ToolCat     string // tool-level Capability.Category
	SearchHints []string
	Exposure    agent.ExposurePolicy
	SafetyLevel string
	ReadOnly    bool
	Activity    string
}

// SearchResult holds a single search hit with its relevance score.
type SearchResult struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Score       float64 `json:"score"`
	MatchField  string  `json:"match_field"`            // which field matched best
	Activity    string  `json:"activity,omitempty"`      // activity kind (read, write, execute, etc.)
}

// NewSearchIndex builds a SearchIndex from the catalog's searchable entries.
func NewSearchIndex(catalog *Catalog) *SearchIndex {
	idx := &SearchIndex{}
	idx.buildFromEntries(catalog.SearchableEntries())
	return idx
}

// Rebuild replaces the index contents from the catalog's current searchable entries.
func (idx *SearchIndex) Rebuild(catalog *Catalog) {
	idx.buildFromEntries(catalog.SearchableEntries())
}

func (idx *SearchIndex) buildFromEntries(entries []ToolEntry) {
	built := make([]searchEntry, 0, len(entries))
	for _, e := range entries {
		cap := e.Tool.Capability
		se := searchEntry{
			Name:        e.Tool.Name,
			Aliases:     cap.Aliases,
			Description: e.Tool.Description,
			CatalogCat:  e.Category,
			ToolCat:     cap.Category,
			SearchHints: cap.SearchHints,
			Exposure:    cap.Exposure,
			SafetyLevel: e.Tool.SafetyLevel.String(),
			ReadOnly:    cap.ReadOnly,
			Activity:    string(cap.Activity),
		}
		built = append(built, se)
	}
	idx.entries = built
}

// Search returns tools matching the query, ranked by weighted scoring.
// Multi-token queries sum scores across all tokens.
// Results are sorted by score descending; ties broken by name ascending.
// A limit of 0 or negative returns all matches.
func (idx *SearchIndex) Search(query string, limit int) []SearchResult {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	var results []SearchResult
	for i := range idx.entries {
		score, bestField := idx.scoreEntry(&idx.entries[i], tokens)
		if score <= 0 {
			continue
		}
		results = append(results, SearchResult{
			Name:        idx.entries[i].Name,
			Description: idx.entries[i].Description,
			Category:    idx.entries[i].CatalogCat,
			Score:       score,
			MatchField:  bestField,
			Activity:    idx.entries[i].Activity,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Name < results[j].Name
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

// scoreEntry computes the total score for an entry across all query tokens.
// It returns the total score and the name of the field that contributed the
// highest single-token score (for MatchField).
func (idx *SearchIndex) scoreEntry(se *searchEntry, tokens []string) (float64, string) {
	var total float64
	var bestField string
	var bestFieldScore float64

	nameLower := strings.ToLower(se.Name)
	descLower := strings.ToLower(se.Description)
	catalogCatLower := strings.ToLower(se.CatalogCat)
	toolCatLower := strings.ToLower(se.ToolCat)
	activityLower := strings.ToLower(se.Activity)

	aliasesLower := make([]string, len(se.Aliases))
	for i, a := range se.Aliases {
		aliasesLower[i] = strings.ToLower(a)
	}

	hintsLower := make([]string, len(se.SearchHints))
	for i, h := range se.SearchHints {
		hintsLower[i] = strings.ToLower(h)
	}

	for _, token := range tokens {
		tokenScore, field := scoreToken(token, nameLower, descLower, catalogCatLower, toolCatLower, activityLower, aliasesLower, hintsLower)
		total += tokenScore
		if tokenScore > bestFieldScore {
			bestFieldScore = tokenScore
			bestField = field
		}
	}

	return total, bestField
}

const (
	weightExactName    = 10.0
	weightPrefixName   = 8.0
	weightExactAlias   = 7.0
	weightPrefixAlias  = 5.0
	weightSearchHint   = 4.0
	weightCategory     = 3.0
	weightDescription  = 2.0
	weightActivity     = 1.0
)

// scoreToken returns the highest score a single token achieves against an entry,
// and the name of the field that produced it.
func scoreToken(
	token string,
	nameLower, descLower, catalogCatLower, toolCatLower, activityLower string,
	aliasesLower, hintsLower []string,
) (float64, string) {
	// Check fields from highest weight to lowest; return on first match
	// to get the best possible score for this token.

	// 1. Exact name match
	if nameLower == token {
		return weightExactName, "name"
	}

	// 2. Name prefix match
	if strings.HasPrefix(nameLower, token) {
		return weightPrefixName, "name"
	}

	// 3. Alias exact match
	for _, alias := range aliasesLower {
		if alias == token {
			return weightExactAlias, "alias"
		}
	}

	// 4. Alias prefix match
	for _, alias := range aliasesLower {
		if strings.HasPrefix(alias, token) {
			return weightPrefixAlias, "alias"
		}
	}

	// 5. SearchHints contain token
	for _, hint := range hintsLower {
		if strings.Contains(hint, token) {
			return weightSearchHint, "search_hint"
		}
	}

	// 6. Category match (catalog or tool)
	if catalogCatLower == token || toolCatLower == token {
		return weightCategory, "category"
	}

	// 7. Description substring
	if strings.Contains(descLower, token) {
		return weightDescription, "description"
	}

	// 8. Activity kind match
	if activityLower != "" && activityLower == token {
		return weightActivity, "activity"
	}

	return 0, ""
}

// tokenize splits the query into lowercase tokens, filtering out empty strings.
func tokenize(query string) []string {
	raw := strings.Fields(strings.ToLower(query))
	tokens := make([]string, 0, len(raw))
	for _, t := range raw {
		if t != "" {
			tokens = append(tokens, t)
		}
	}
	return tokens
}
