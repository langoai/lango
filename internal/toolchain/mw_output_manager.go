package toolchain

import (
	"context"
	"encoding/json"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/tooloutput"
	"github.com/langoai/lango/internal/types"
)

const (
	defaultTokenBudget = 2000
	defaultHeadRatio   = 0.7
	defaultTailRatio   = 0.3

	tierSmall  = "small"
	tierMedium = "medium"
	tierLarge  = "large"
)

// OutputStorer is the subset of tooloutput.OutputStore used by the middleware.
type OutputStorer interface {
	Store(toolName, content string) string
}

// outputMeta holds metadata about the output processing.
type outputMeta struct {
	OriginalTokens int
	Tier           string
	ContentType    tooloutput.ContentType
	Compressed     bool
	StoredRef      *string
}

// WithOutputManager returns a middleware that manages tool output based on token budgets.
// It classifies output into tiers (small/medium/large) and applies content-aware compression
// when output exceeds the configured token budget. An optional OutputStorer stores large
// outputs for later retrieval via tool_output_get.
func WithOutputManager(cfg config.OutputManagerConfig, store ...OutputStorer) Middleware {
	enabled := cfg.Enabled == nil || *cfg.Enabled
	budget := cfg.TokenBudget
	if budget <= 0 {
		budget = defaultTokenBudget
	}
	headRatio := cfg.HeadRatio
	if headRatio <= 0 {
		headRatio = defaultHeadRatio
	}
	tailRatio := cfg.TailRatio
	if tailRatio <= 0 {
		tailRatio = defaultTailRatio
	}

	var outputStore OutputStorer
	if len(store) > 0 {
		outputStore = store[0]
	}

	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			result, err := next(ctx, params)
			if err != nil {
				return result, err
			}
			if !enabled {
				return result, nil
			}

			return processOutput(result, budget, headRatio, tailRatio, tool.Name, outputStore), nil
		}
	}
}

// processOutput applies token-based tier processing to the tool result.
func processOutput(result interface{}, budget int, headRatio, tailRatio float64, toolName string, store OutputStorer) interface{} {
	text, isString := resultToText(result)
	if text == "" {
		return result
	}

	// Fast path: skip full token estimation for obviously small outputs.
	// EstimateTokens uses ~4 chars/token for ASCII, so len < budget*4 is always under budget.
	if len(text) < budget*4 {
		ct := tooloutput.DetectContentType(text)
		tokens := types.EstimateTokens(text)
		return injectMeta(result, isString, outputMeta{
			OriginalTokens: tokens,
			Tier:           tierSmall,
			ContentType:    ct,
			Compressed:     false,
		})
	}

	tokens := types.EstimateTokens(text)
	ct := tooloutput.DetectContentType(text)

	switch {
	case tokens <= budget:
		// Small tier: pass through with metadata.
		return injectMeta(result, isString, outputMeta{
			OriginalTokens: tokens,
			Tier:           tierSmall,
			ContentType:    ct,
			Compressed:     false,
		})

	case tokens <= 3*budget:
		// Medium tier: content-aware compress.
		compressed := tooloutput.Compress(text, ct, budget, headRatio, tailRatio)
		logging.App().Infow("output manager compressed medium output",
			"tool", toolName,
			"originalTokens", tokens,
			"budget", budget)
		return injectMeta(compressed, true, outputMeta{
			OriginalTokens: tokens,
			Tier:           tierMedium,
			ContentType:    ct,
			Compressed:     true,
		})

	default:
		// Large tier: aggressive content-aware compress + store for retrieval.
		aggressiveBudget := budget / 2
		if aggressiveBudget < 1 {
			aggressiveBudget = 1
		}
		compressed := tooloutput.Compress(text, ct, aggressiveBudget, headRatio, tailRatio)
		logging.App().Warnw("output manager aggressively compressed large output",
			"tool", toolName,
			"originalTokens", tokens,
			"budget", budget)

		meta := outputMeta{
			OriginalTokens: tokens,
			Tier:           tierLarge,
			ContentType:    ct,
			Compressed:     true,
		}

		// Store full output for later retrieval if store is available.
		if store != nil {
			ref := store.Store(toolName, text)
			meta.StoredRef = &ref
		}

		return injectMeta(compressed, true, meta)
	}
}

// resultToText converts a tool result to its text representation.
// Returns the text and whether the original result was a string.
func resultToText(result interface{}) (string, bool) {
	switch v := result.(type) {
	case string:
		return v, true
	case nil:
		return "", false
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", false
		}
		return string(data), false
	}
}

// injectMeta adds _meta to the result.
// If the result was a string, wraps it as {"content": "...", "_meta": {...}}.
// If the result was a map, injects _meta directly.
func injectMeta(result interface{}, isString bool, meta outputMeta) interface{} {
	metaMap := map[string]interface{}{
		"originalTokens": meta.OriginalTokens,
		"tier":           meta.Tier,
		"contentType":    string(meta.ContentType),
		"compressed":     meta.Compressed,
	}
	if meta.StoredRef != nil {
		metaMap["storedRef"] = *meta.StoredRef
	} else if meta.Tier == tierLarge {
		metaMap["storedRef"] = nil
	}

	if isString {
		text, _ := result.(string)
		return map[string]interface{}{
			"content": text,
			"_meta":   metaMap,
		}
	}

	// If the result is a map, inject _meta directly.
	if m, ok := result.(map[string]interface{}); ok {
		m["_meta"] = metaMap
		return m
	}

	// Fallback for non-string, non-map results: marshal to map then inject.
	data, err := json.Marshal(result)
	if err != nil {
		return result
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		// Cannot convert to map; wrap the marshaled JSON as content.
		return map[string]interface{}{
			"content": string(data),
			"_meta":   metaMap,
		}
	}
	m["_meta"] = metaMap
	return m
}
