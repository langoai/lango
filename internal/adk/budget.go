package adk

import (
	"fmt"
	"math"
	"strings"
)

// SectionAllocation defines the ratio of available context budget allocated to each section.
// All values must sum to 1.0 (within tolerance of 0.001).
type SectionAllocation struct {
	Knowledge  float64
	RAG        float64
	Memory     float64
	RunSummary float64
	Headroom   float64
}

// DefaultAllocation returns the default section allocation ratios.
func DefaultAllocation() SectionAllocation {
	return SectionAllocation{
		Knowledge:  0.30,
		RAG:        0.25,
		Memory:     0.25,
		RunSummary: 0.10,
		Headroom:   0.10,
	}
}

// sum returns the total of all allocation ratios.
func (a SectionAllocation) sum() float64 {
	return a.Knowledge + a.RAG + a.Memory + a.RunSummary + a.Headroom
}

// SectionBudgets holds computed per-section token budgets.
// A value of 0 means unlimited (no budget enforcement).
// Degraded is true when available budget was zero or negative.
type SectionBudgets struct {
	Knowledge  int
	RAG        int
	Memory     int
	RunSummary int
	Degraded   bool
}

// ContextBudgetManager allocates available context window tokens across prompt sections.
type ContextBudgetManager struct {
	modelWindow      int
	responseReserve  int
	basePromptTokens int
	allocation       SectionAllocation
}

// NewContextBudgetManager creates a budget manager with validated allocation.
// The allocation sum must equal 1.0 (within 0.001 tolerance).
// responseReserve is clamped to [1024, 25% of modelWindow].
func NewContextBudgetManager(modelWindow, responseReserve, basePromptTokens int, alloc SectionAllocation) (*ContextBudgetManager, error) {
	if math.Abs(alloc.sum()-1.0) > 0.001 {
		return nil, fmt.Errorf("allocation sum must equal 1.0, got %.4f", alloc.sum())
	}

	// Clamp response reserve.
	if responseReserve <= 0 {
		responseReserve = 4096
	}
	minReserve := 1024
	maxReserve := modelWindow / 4
	if maxReserve < minReserve {
		maxReserve = minReserve
	}
	if responseReserve < minReserve {
		responseReserve = minReserve
	}
	if responseReserve > maxReserve {
		responseReserve = maxReserve
	}

	return &ContextBudgetManager{
		modelWindow:      modelWindow,
		responseReserve:  responseReserve,
		basePromptTokens: basePromptTokens,
		allocation:       alloc,
	}, nil
}

// SectionBudgets computes per-section token budgets.
// Returns zero budgets (unlimited) when available budget is zero or negative (degradation).
// Check Degraded field on the result to know if budget enforcement was skipped.
func (bm *ContextBudgetManager) SectionBudgets() SectionBudgets {
	available := bm.modelWindow - bm.responseReserve - bm.basePromptTokens
	if available <= 0 {
		return SectionBudgets{Degraded: true} // All zeros = unlimited (degradation).
	}

	return SectionBudgets{
		Knowledge:  int(float64(available) * bm.allocation.Knowledge),
		RAG:        int(float64(available) * bm.allocation.RAG),
		Memory:     int(float64(available) * bm.allocation.Memory),
		RunSummary: int(float64(available) * bm.allocation.RunSummary),
	}
}

// ModelWindow returns the configured model window size.
func (bm *ContextBudgetManager) ModelWindow() int {
	return bm.modelWindow
}

// SectionTokens represents measured token counts per section before truncation.
// Used as input to ReallocateBudgets for empty-section redistribution.
type SectionTokens struct {
	Knowledge  int
	RAG        int
	Memory     int
	RunSummary int
}

// ReallocateBudgets computes per-section budgets with empty-section redistribution.
// Sections with measured token count of 0 donate their entire budget proportionally
// to sections that have content. Non-empty sections keep their full initial budget
// plus a proportional share of the surplus. Headroom is never redistributed.
//
// No recursive redistribution: surplus is distributed once. Excess in non-empty
// sections is simply unused.
//
// All sections empty: returns all-zero budgets with Degraded=false.
func (bm *ContextBudgetManager) ReallocateBudgets(measured SectionTokens) SectionBudgets {
	base := bm.SectionBudgets()
	if base.Degraded {
		return base
	}

	// Fast path: all sections have content, no reallocation needed.
	if measured.Knowledge > 0 && measured.RAG > 0 && measured.Memory > 0 && measured.RunSummary > 0 {
		return base
	}

	budgets := [4]int{base.Knowledge, base.RAG, base.Memory, base.RunSummary}
	ratios := [4]float64{bm.allocation.Knowledge, bm.allocation.RAG, bm.allocation.Memory, bm.allocation.RunSummary}
	tokens := [4]int{measured.Knowledge, measured.RAG, measured.Memory, measured.RunSummary}

	var surplus int
	var presentRatioSum float64
	for i := 0; i < 4; i++ {
		if tokens[i] == 0 {
			surplus += budgets[i]
			budgets[i] = 0
		} else {
			presentRatioSum += ratios[i]
		}
	}

	// No empty sections → no reallocation.
	if surplus == 0 {
		return base
	}

	// All sections empty → all-zero budgets, no context for this turn.
	if presentRatioSum == 0 {
		return SectionBudgets{}
	}

	// Redistribute surplus proportionally to present sections.
	for i := 0; i < 4; i++ {
		if tokens[i] > 0 {
			budgets[i] += int(float64(surplus) * ratios[i] / presentRatioSum)
		}
	}

	return SectionBudgets{
		Knowledge:  budgets[0],
		RAG:        budgets[1],
		Memory:     budgets[2],
		RunSummary: budgets[3],
	}
}

// modelWindowRegistry maps model name prefixes to context window sizes in tokens.
var modelWindowRegistry = map[string]int{
	// Google Gemini
	"gemini-2.0":  1000000,
	"gemini-1.5":  2000000,
	"gemini-1.0":  32000,
	"gemini-pro":  32000,
	"gemini-nano": 8000,

	// Anthropic Claude
	"claude-opus":   200000,
	"claude-sonnet": 200000,
	"claude-haiku":  200000,
	"claude-3":      200000,
	"claude-4":      200000,

	// OpenAI GPT
	"gpt-4o":       128000,
	"gpt-4-turbo":  128000,
	"gpt-4-0125":   128000,
	"gpt-4-1106":   128000,
	"gpt-4":        8192,
	"gpt-3.5":      16385,
	"o1":           200000,
	"o3":           200000,

	// Meta Llama
	"llama-3.3": 128000,
	"llama-3.2": 128000,
	"llama-3.1": 128000,
	"llama3":    128000,
	"llama-3":   128000,

	// Mistral
	"mistral-large":  128000,
	"mistral-medium": 32000,
	"mistral-small":  32000,

	// Local / other
	"deepseek": 128000,
	"qwen":     128000,
	"phi":      128000,
}

const defaultModelWindow = 128000

// LookupModelWindow returns the context window size for the given model name.
// Matches by longest prefix. Returns defaultModelWindow (128k) for unknown models.
func LookupModelWindow(modelName string) int {
	modelName = strings.ToLower(modelName)

	// Try exact prefix matches from longest to shortest.
	bestMatch := ""
	bestWindow := 0
	for prefix, window := range modelWindowRegistry {
		if strings.HasPrefix(modelName, prefix) && len(prefix) > len(bestMatch) {
			bestMatch = prefix
			bestWindow = window
		}
	}
	if bestMatch != "" {
		return bestWindow
	}
	return defaultModelWindow
}
