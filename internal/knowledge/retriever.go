package knowledge

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// ToolRegistryProvider supplies available tool information.
type ToolRegistryProvider interface {
	ListTools() []ToolDescriptor
	SearchTools(query string, limit int) []ToolDescriptor
}

// RuntimeContextProvider supplies current session/system state.
type RuntimeContextProvider interface {
	GetRuntimeContext() RuntimeContext
}

// SkillInfo describes a single skill for context retrieval.
type SkillInfo struct {
	Name        string
	Description string
	Type        string
}

// SkillProvider supplies active skill information.
type SkillProvider interface {
	ListActiveSkillInfos(ctx context.Context) ([]SkillInfo, error)
}

// ContextRetriever searches the context layers and assembles augmented prompts.
type ContextRetriever struct {
	store           *Store
	maxPerLayer     int
	logger          *zap.SugaredLogger
	toolProvider    ToolRegistryProvider
	runtimeProvider RuntimeContextProvider
	skillProvider   SkillProvider
	inquiryProvider InquiryProvider
}

// NewContextRetriever creates a new context retriever.
func NewContextRetriever(store *Store, maxPerLayer int, logger *zap.SugaredLogger) *ContextRetriever {
	if maxPerLayer <= 0 {
		maxPerLayer = 5
	}
	return &ContextRetriever{
		store:       store,
		maxPerLayer: maxPerLayer,
		logger:      logger,
	}
}

// WithToolRegistry sets the tool registry provider.
func (r *ContextRetriever) WithToolRegistry(p ToolRegistryProvider) *ContextRetriever {
	r.toolProvider = p
	return r
}

// WithRuntimeContext sets the runtime context provider.
func (r *ContextRetriever) WithRuntimeContext(p RuntimeContextProvider) *ContextRetriever {
	r.runtimeProvider = p
	return r
}

// WithSkillProvider sets the skill provider for context retrieval.
func (r *ContextRetriever) WithSkillProvider(p SkillProvider) *ContextRetriever {
	r.skillProvider = p
	return r
}

// WithInquiryProvider sets the inquiry provider for proactive librarian context.
func (r *ContextRetriever) WithInquiryProvider(p InquiryProvider) *ContextRetriever {
	r.inquiryProvider = p
	return r
}

// Retrieve searches the requested context layers and returns relevant items.
func (r *ContextRetriever) Retrieve(ctx context.Context, req RetrievalRequest) (*RetrievalResult, error) {
	result := &RetrievalResult{
		Items: make(map[ContextLayer][]ContextItem),
	}

	maxPerLayer := req.MaxPerLayer
	if maxPerLayer <= 0 {
		maxPerLayer = r.maxPerLayer
	}

	layers := req.Layers
	if len(layers) == 0 {
		layers = []ContextLayer{
			LayerUserKnowledge,
			LayerSkillPatterns,
			LayerExternalKnowledge,
			LayerAgentLearnings,
		}
	}

	keywords := extractKeywords(req.Query)
	if len(keywords) == 0 {
		return result, nil
	}
	searchQuery := strings.Join(keywords, " ")

	for _, layer := range layers {
		var items []ContextItem
		var err error

		switch layer {
		case LayerUserKnowledge:
			items, err = r.retrieveKnowledge(ctx, searchQuery, maxPerLayer)
		case LayerSkillPatterns:
			items, err = r.retrieveSkills(ctx, searchQuery, maxPerLayer)
		case LayerExternalKnowledge:
			items, err = r.retrieveExternalRefs(ctx, searchQuery, maxPerLayer)
		case LayerAgentLearnings:
			items, err = r.retrieveLearnings(ctx, searchQuery, maxPerLayer)
		case LayerToolRegistry:
			items = r.retrieveTools(searchQuery, maxPerLayer)
		case LayerRuntimeContext:
			items = r.retrieveRuntimeContext()
		case LayerPendingInquiries:
			items, err = r.retrievePendingInquiries(ctx, req.SessionKey, maxPerLayer)
		}

		if err != nil {
			r.logger.Warnw("context retrieval error", "layer", layer, "error", err)
			continue
		}

		if len(items) > 0 {
			result.Items[layer] = items
			result.TotalItems += len(items)
		}
	}

	return result, nil
}

// AssemblePrompt builds an augmented system prompt from base prompt and retrieved context.
func (r *ContextRetriever) AssemblePrompt(basePrompt string, result *RetrievalResult) string {
	if result == nil || result.TotalItems == 0 {
		return basePrompt
	}

	var b strings.Builder
	b.WriteString(basePrompt)

	if items, ok := result.Items[LayerRuntimeContext]; ok && len(items) > 0 {
		b.WriteString("\n\n## Runtime Context\n")
		for _, item := range items {
			b.WriteString(fmt.Sprintf("- %s\n", item.Content))
		}
	}

	if items, ok := result.Items[LayerToolRegistry]; ok && len(items) > 0 {
		b.WriteString("\n\n## Available Tools\n")
		for _, item := range items {
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", item.Key, item.Content))
		}
	}

	if items, ok := result.Items[LayerUserKnowledge]; ok && len(items) > 0 {
		b.WriteString("\n\n## User Knowledge\n")
		for _, item := range items {
			b.WriteString(fmt.Sprintf("- [%s] %s: %s\n", item.Category, item.Key, item.Content))
		}
	}

	if items, ok := result.Items[LayerAgentLearnings]; ok && len(items) > 0 {
		b.WriteString("\n\n## Known Solutions\n")
		for _, item := range items {
			b.WriteString(fmt.Sprintf("- %s\n", item.Content))
		}
	}

	if items, ok := result.Items[LayerSkillPatterns]; ok && len(items) > 0 {
		b.WriteString("\n\n## Available Skills\n")
		b.WriteString("**Note:** Prefer built-in tools over skills. Use skills only when no built-in tool provides the needed functionality.\n")
		for _, item := range items {
			b.WriteString(fmt.Sprintf("- %s: %s\n", item.Key, item.Content))
		}
	}

	if items, ok := result.Items[LayerExternalKnowledge]; ok && len(items) > 0 {
		b.WriteString("\n\n## External References\n")
		for _, item := range items {
			b.WriteString(fmt.Sprintf("- %s (%s): %s\n", item.Key, item.Source, item.Content))
		}
	}

	if items, ok := result.Items[LayerPendingInquiries]; ok && len(items) > 0 {
		b.WriteString("\n\n## Pending Knowledge Inquiries\n")
		b.WriteString("Consider weaving ONE of these questions naturally into your response:\n")
		for _, item := range items {
			if item.Source != "" {
				b.WriteString(fmt.Sprintf("- [%s] %s (context: %s)\n", item.Key, item.Content, item.Source))
			} else {
				b.WriteString(fmt.Sprintf("- [%s] %s\n", item.Key, item.Content))
			}
		}
	}

	return b.String()
}

func (r *ContextRetriever) retrieveTools(query string, limit int) []ContextItem {
	if r.toolProvider == nil {
		return nil
	}
	tools := r.toolProvider.SearchTools(query, limit)
	items := make([]ContextItem, 0, len(tools))
	for _, t := range tools {
		items = append(items, ContextItem{
			Layer:   LayerToolRegistry,
			Key:     t.Name,
			Content: t.Description,
		})
	}
	return items
}

func (r *ContextRetriever) retrieveRuntimeContext() []ContextItem {
	if r.runtimeProvider == nil {
		return nil
	}
	rc := r.runtimeProvider.GetRuntimeContext()
	content := fmt.Sprintf(
		"Session: %s | Channel: %s | Tools: %d | Encryption: %v | Knowledge: %v | Memory: %v",
		rc.SessionKey, rc.ChannelType, rc.ActiveToolCount,
		rc.EncryptionEnabled, rc.KnowledgeEnabled, rc.MemoryEnabled,
	)
	return []ContextItem{{
		Layer:   LayerRuntimeContext,
		Key:     "session-state",
		Content: content,
	}}
}

func (r *ContextRetriever) retrieveKnowledge(ctx context.Context, query string, limit int) ([]ContextItem, error) {
	entries, err := r.store.SearchKnowledge(ctx, query, "", limit)
	if err != nil {
		return nil, err
	}

	items := make([]ContextItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, ContextItem{
			Layer:    LayerUserKnowledge,
			Key:      e.Key,
			Content:  e.Content,
			Category: string(e.Category),
			Source:   e.Source,
		})
	}
	return items, nil
}

func (r *ContextRetriever) retrieveSkills(ctx context.Context, query string, limit int) ([]ContextItem, error) {
	if r.skillProvider == nil {
		return nil, nil
	}

	skills, err := r.skillProvider.ListActiveSkillInfos(ctx)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var items []ContextItem
	for _, sk := range skills {
		if len(items) >= limit {
			break
		}
		nameLower := strings.ToLower(sk.Name)
		descLower := strings.ToLower(sk.Description)
		if strings.Contains(nameLower, queryLower) || strings.Contains(descLower, queryLower) {
			items = append(items, ContextItem{
				Layer:    LayerSkillPatterns,
				Key:      sk.Name,
				Content:  sk.Description,
				Category: sk.Type,
			})
		}
	}
	return items, nil
}

func (r *ContextRetriever) retrieveExternalRefs(ctx context.Context, query string, limit int) ([]ContextItem, error) {
	refs, err := r.store.SearchExternalRefs(ctx, query)
	if err != nil {
		return nil, err
	}

	items := make([]ContextItem, 0, len(refs))
	for i, ref := range refs {
		if i >= limit {
			break
		}
		items = append(items, ContextItem{
			Layer:   LayerExternalKnowledge,
			Key:     ref.Name,
			Content: ref.Summary,
			Source:  ref.Location,
		})
	}
	return items, nil
}

func (r *ContextRetriever) retrieveLearnings(ctx context.Context, query string, limit int) ([]ContextItem, error) {
	learnings, err := r.store.SearchLearnings(ctx, query, "", limit)
	if err != nil {
		return nil, err
	}

	var items []ContextItem
	for _, l := range learnings {
		content := l.Trigger
		if l.Fix != "" {
			content = fmt.Sprintf("When '%s' occurs: %s", l.Trigger, l.Fix)
		}
		items = append(items, ContextItem{
			Layer:    LayerAgentLearnings,
			Key:      l.Trigger,
			Content:  content,
			Category: string(l.Category),
		})
	}
	return items, nil
}

func (r *ContextRetriever) retrievePendingInquiries(ctx context.Context, sessionKey string, limit int) ([]ContextItem, error) {
	if r.inquiryProvider == nil {
		return nil, nil
	}
	return r.inquiryProvider.PendingInquiryItems(ctx, sessionKey, limit)
}

// Common English stop words to filter from queries.
var _stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true,
	"did": true, "will": true, "would": true, "could": true, "should": true,
	"may": true, "might": true, "can": true, "shall": true,
	"to": true, "of": true, "in": true, "for": true, "on": true,
	"with": true, "at": true, "by": true, "from": true, "as": true,
	"into": true, "about": true, "between": true,
	"and": true, "or": true, "but": true, "not": true, "no": true,
	"it": true, "its": true, "this": true, "that": true, "these": true,
	"those": true, "my": true, "your": true, "his": true, "her": true,
	"our": true, "their": true, "what": true, "which": true, "who": true,
	"how": true, "when": true, "where": true, "why": true,
	"i": true, "me": true, "we": true, "you": true, "he": true, "she": true, "they": true,
}

const (
	_maxSearchKeywords = 5
	_maxKeywordLength  = 50
)

// _reNonAlphaNum matches characters that are not alphanumeric, hyphens, or underscores.
var _reNonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9\-_]`)

// extractKeywords extracts meaningful keywords from a query string.
// Results are limited to _maxSearchKeywords items, each at most _maxKeywordLength chars,
// to prevent SQLite LIKE pattern complexity issues.
func extractKeywords(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	var keywords []string
	for _, w := range words {
		// Remove punctuation
		w = strings.Trim(w, ".,!?;:'\"()[]{}")
		if len(w) < 2 {
			continue
		}
		if _stopWords[w] {
			continue
		}
		w = sanitizeKeyword(w)
		if len(w) < 2 {
			continue
		}
		keywords = append(keywords, w)
		if len(keywords) >= _maxSearchKeywords {
			break
		}
	}
	return keywords
}

// sanitizeKeyword strips non-alphanumeric characters and truncates to _maxKeywordLength.
func sanitizeKeyword(w string) string {
	w = _reNonAlphaNum.ReplaceAllString(w, "")
	if len(w) > _maxKeywordLength {
		w = w[:_maxKeywordLength]
	}
	return w
}
