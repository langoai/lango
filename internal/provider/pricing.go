package provider

import "strings"

// ModelPrice describes per-million-token pricing in USD for a model.
type ModelPrice struct {
	InputPerMillion  float64
	OutputPerMillion float64
}

// modelPrices is a static price table for the primary supported models.
// Prices are USD per million tokens. Entries match by model-name prefix
// (case-insensitive) so versioned IDs like "claude-opus-4-6-20251119" hit
// the "claude-opus" prefix. Unknown models return ({}, false).
//
// Update this table as provider pricing changes; stale entries are safer to
// remove than to overestimate cost.
var modelPrices = []struct {
	prefix string
	price  ModelPrice
}{
	// Anthropic Claude — public list pricing as of 2025.
	{"claude-opus-4", ModelPrice{InputPerMillion: 15.0, OutputPerMillion: 75.0}},
	{"claude-sonnet-4", ModelPrice{InputPerMillion: 3.0, OutputPerMillion: 15.0}},
	{"claude-haiku-4", ModelPrice{InputPerMillion: 0.80, OutputPerMillion: 4.0}},
	{"claude-3-5-sonnet", ModelPrice{InputPerMillion: 3.0, OutputPerMillion: 15.0}},

	// Google Gemini.
	{"gemini-2.5-pro", ModelPrice{InputPerMillion: 1.25, OutputPerMillion: 10.0}},
	{"gemini-2.5-flash", ModelPrice{InputPerMillion: 0.30, OutputPerMillion: 2.50}},
	{"gemini-2.0-flash", ModelPrice{InputPerMillion: 0.10, OutputPerMillion: 0.40}},

	// OpenAI GPT.
	{"gpt-4o", ModelPrice{InputPerMillion: 2.50, OutputPerMillion: 10.0}},
}

// PriceFor returns the price entry for the given model name. The lookup is
// case-insensitive prefix match. Returns (zero, false) when no entry matches.
func PriceFor(model string) (ModelPrice, bool) {
	if model == "" {
		return ModelPrice{}, false
	}
	lower := strings.ToLower(model)
	for _, entry := range modelPrices {
		if strings.HasPrefix(lower, entry.prefix) {
			return entry.price, true
		}
	}
	return ModelPrice{}, false
}

// EstimateCostUSD computes the estimated cost in USD for the given model and
// token counts. Returns 0 when the model has no pricing entry.
func EstimateCostUSD(model string, inputTokens, outputTokens int) float64 {
	price, ok := PriceFor(model)
	if !ok {
		return 0
	}
	return (float64(inputTokens)*price.InputPerMillion + float64(outputTokens)*price.OutputPerMillion) / 1_000_000
}
