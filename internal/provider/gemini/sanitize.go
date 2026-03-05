package gemini

import "google.golang.org/genai"

// sanitizeContents ensures the content sequence is valid for the Gemini API.
//
// Gemini enforces strict turn-ordering rules:
//  1. No two consecutive same-role turns.
//  2. A model turn with FunctionCall must be immediately followed by a user/function
//     turn with FunctionResponse.
//  3. The first content must be role "user".
//  4. Orphaned FunctionResponse without a preceding FunctionCall is removed.
//
// This function merges, reorders, and patches the content slice so that these
// invariants hold regardless of upstream data quality. Already-valid sequences
// pass through with minimal overhead (single O(n) scan).
func sanitizeContents(contents []*genai.Content) []*genai.Content {
	if len(contents) == 0 {
		return contents
	}

	// Step 1: Remove orphaned FunctionResponse at the start (no preceding FunctionCall).
	// Must run before merging so orphan parts are not folded into normal user turns.
	contents = dropLeadingOrphanedFunctionResponses(contents)

	// Step 2: Merge consecutive same-role turns.
	merged := mergeConsecutiveRoles(contents)

	// Step 3: Ensure first turn is "user". If it starts with "model", prepend
	// a synthetic user turn so the API accepts the sequence.
	if len(merged) > 0 && merged[0].Role == "model" {
		synthetic := &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{{Text: "[continue]"}},
		}
		merged = append([]*genai.Content{synthetic}, merged...)
	}

	// Step 4: Ensure every model+FunctionCall is followed by a user/function
	// turn containing FunctionResponse. Insert synthetic responses where missing.
	merged = ensureFunctionResponsePairs(merged)

	// Step 5: Final merge pass — step 4 may have inserted synthetic user turns
	// adjacent to existing user turns.
	merged = mergeConsecutiveRoles(merged)

	return merged
}

// mergeConsecutiveRoles walks the slice and merges adjacent contents that share
// the same role into a single content by concatenating their Parts.
func mergeConsecutiveRoles(contents []*genai.Content) []*genai.Content {
	if len(contents) == 0 {
		return nil
	}

	result := make([]*genai.Content, 0, len(contents))
	current := cloneContent(contents[0])

	for i := 1; i < len(contents); i++ {
		if contents[i].Role == current.Role {
			current.Parts = append(current.Parts, contents[i].Parts...)
		} else {
			result = append(result, current)
			current = cloneContent(contents[i])
		}
	}
	result = append(result, current)
	return result
}

// dropLeadingOrphanedFunctionResponses removes user/function turns at the start
// of the sequence that contain only FunctionResponse parts (no preceding
// model+FunctionCall to match them).
func dropLeadingOrphanedFunctionResponses(contents []*genai.Content) []*genai.Content {
	for len(contents) > 0 && containsOnlyFunctionResponses(contents[0]) {
		contents = contents[1:]
	}
	return contents
}

// ensureFunctionResponsePairs scans for model turns containing FunctionCall
// parts. If the immediately following turn does not contain a matching
// FunctionResponse, a synthetic one is inserted.
func ensureFunctionResponsePairs(contents []*genai.Content) []*genai.Content {
	result := make([]*genai.Content, 0, len(contents))

	for i := 0; i < len(contents); i++ {
		result = append(result, contents[i])

		if !hasFunctionCallParts(contents[i]) {
			continue
		}

		// Check whether the next turn has FunctionResponse.
		hasResponse := i+1 < len(contents) && hasFunctionResponseParts(contents[i+1])
		if hasResponse {
			continue
		}

		// Insert synthetic FunctionResponse for each FunctionCall.
		var responseParts []*genai.Part
		for _, p := range contents[i].Parts {
			if p.FunctionCall != nil {
				responseParts = append(responseParts, &genai.Part{
					FunctionResponse: &genai.FunctionResponse{
						Name:     p.FunctionCall.Name,
						Response: map[string]any{"result": "[no response available]"},
					},
				})
			}
		}
		if len(responseParts) > 0 {
			result = append(result, &genai.Content{
				Role:  "user",
				Parts: responseParts,
			})
		}
	}

	return result
}

// hasFunctionCallParts reports whether the content contains at least one FunctionCall part.
func hasFunctionCallParts(c *genai.Content) bool {
	if c == nil || c.Role != "model" {
		return false
	}
	for _, p := range c.Parts {
		if p.FunctionCall != nil {
			return true
		}
	}
	return false
}

// hasFunctionResponseParts reports whether the content contains at least one FunctionResponse part.
func hasFunctionResponseParts(c *genai.Content) bool {
	if c == nil {
		return false
	}
	for _, p := range c.Parts {
		if p.FunctionResponse != nil {
			return true
		}
	}
	return false
}

// containsOnlyFunctionResponses reports whether the content has only FunctionResponse parts
// (no text or other content). Such turns are orphaned when truncation removed the
// preceding model+FunctionCall.
func containsOnlyFunctionResponses(c *genai.Content) bool {
	if c == nil || len(c.Parts) == 0 {
		return false
	}
	for _, p := range c.Parts {
		if p.FunctionResponse == nil {
			return false
		}
	}
	return true
}

// cloneContent creates a shallow copy of a Content so that mutations (e.g.
// appending Parts) do not affect the original.
func cloneContent(c *genai.Content) *genai.Content {
	parts := make([]*genai.Part, len(c.Parts))
	copy(parts, c.Parts)
	return &genai.Content{
		Role:  c.Role,
		Parts: parts,
	}
}
