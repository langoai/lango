package tooloutput

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/langoai/lango/internal/types"
)

var (
	logErrorWarnPattern = regexp.MustCompile(`(?i)\b(ERROR|WARN)\b`)

	codeSignaturePattern = regexp.MustCompile(
		`(?m)^.*(func |class |def |type |import |package ).*$`,
	)

	goroutineBoundary = regexp.MustCompile(`(?m)^goroutine \d+`)
)

// CompressJSON compresses JSON by extracting schema + sample + count.
// For arrays: shows first N items + total count.
// For objects: shows keys + truncated values.
func CompressJSON(text string, maxTokens int) string {
	if types.EstimateTokens(text) <= maxTokens {
		return text
	}

	trimmed := strings.TrimSpace(text)

	// Try to parse as array
	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
		return compressJSONArray(arr, maxTokens)
	}

	// Try to parse as object
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
		return compressJSONObject(obj, maxTokens)
	}

	// Parse failed — fallback
	return CompressHeadTail(text, 0.7, 0.3, maxTokens)
}

func compressJSONArray(arr []json.RawMessage, maxTokens int) string {
	total := len(arr)
	sampleCount := 2
	if total <= sampleCount {
		sampleCount = total
	}

	sample := make([]json.RawMessage, sampleCount)
	copy(sample, arr[:sampleCount])

	out, err := json.MarshalIndent(sample, "", "  ")
	if err != nil {
		return fmt.Sprintf("[%d items, marshal error]", total)
	}

	result := string(out)
	if total > sampleCount {
		result += fmt.Sprintf("\n... and %d more items (total: %d)", total-sampleCount, total)
	}

	if types.EstimateTokens(result) > maxTokens {
		return CompressHeadTail(result, 0.7, 0.3, maxTokens)
	}
	return result
}

func compressJSONObject(obj map[string]json.RawMessage, maxTokens int) string {
	truncated := make(map[string]any, len(obj))
	for k, v := range obj {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			if len(s) > 100 {
				s = s[:100] + "..."
			}
			truncated[k] = s
			continue
		}

		var nested any
		if err := json.Unmarshal(v, &nested); err == nil {
			raw := string(v)
			if len(raw) > 100 {
				truncated[k] = json.RawMessage(raw[:100] + "...")
			} else {
				truncated[k] = nested
			}
		} else {
			truncated[k] = string(v)
		}
	}

	out, err := json.MarshalIndent(truncated, "", "  ")
	if err != nil {
		return fmt.Sprintf("{%d keys, marshal error}", len(obj))
	}

	result := string(out)
	if types.EstimateTokens(result) > maxTokens {
		return CompressHeadTail(result, 0.7, 0.3, maxTokens)
	}
	return result
}

// CompressLog compresses log output by extracting ERROR/WARN lines,
// then head/tail of remaining lines, with a summary.
func CompressLog(text string, maxTokens int) string {
	if types.EstimateTokens(text) <= maxTokens {
		return text
	}

	lines := strings.Split(text, "\n")
	var errorLines []string
	for _, line := range lines {
		if logErrorWarnPattern.MatchString(line) {
			errorLines = append(errorLines, line)
		}
	}

	totalLines := len(lines)
	summary := fmt.Sprintf("[log summary: %d total lines, %d error/warn lines]",
		totalLines, len(errorLines))

	if len(errorLines) > 0 {
		extracted := strings.Join(errorLines, "\n")
		result := summary + "\n" + extracted
		if types.EstimateTokens(result) <= maxTokens {
			return result
		}
		// Error/warn lines alone exceed budget — compress them
		return summary + "\n" + CompressHeadTail(extracted, 0.7, 0.3, maxTokens-types.EstimateTokens(summary)-1)
	}

	return summary + "\n" + CompressHeadTail(text, 0.7, 0.3, maxTokens-types.EstimateTokens(summary)-1)
}

// CompressCode compresses code by extracting signatures/imports,
// then head/tail of the body.
func CompressCode(text string, maxTokens int) string {
	if types.EstimateTokens(text) <= maxTokens {
		return text
	}

	matches := codeSignaturePattern.FindAllString(text, -1)
	signatureBlock := ""
	if len(matches) > 0 {
		signatureBlock = "[signatures]\n" + strings.Join(matches, "\n")
	}

	sigTokens := types.EstimateTokens(signatureBlock)
	remaining := maxTokens - sigTokens
	if remaining <= 0 {
		return CompressHeadTail(signatureBlock, 0.7, 0.3, maxTokens)
	}

	body := CompressHeadTail(text, 0.7, 0.3, remaining)
	if signatureBlock == "" {
		return body
	}
	return signatureBlock + "\n\n[body]\n" + body
}

// CompressStackTrace compresses stack traces by keeping the first
// goroutine/thread fully, summarizing the rest.
func CompressStackTrace(text string, maxTokens int) string {
	if types.EstimateTokens(text) <= maxTokens {
		return text
	}

	locs := goroutineBoundary.FindAllStringIndex(text, -1)
	if len(locs) <= 1 {
		return CompressHeadTail(text, 0.7, 0.3, maxTokens)
	}

	// Keep everything up to the second goroutine boundary
	firstBlock := text[:locs[1][0]]
	remaining := len(locs) - 1

	summary := fmt.Sprintf("\n... [%d more goroutines/threads omitted]", remaining)
	result := strings.TrimRight(firstBlock, "\n") + summary

	if types.EstimateTokens(result) > maxTokens {
		return CompressHeadTail(result, 0.7, 0.3, maxTokens)
	}
	return result
}

// CompressHeadTail is the generic fallback compressor.
// Takes headRatio/tailRatio (e.g. 0.7/0.3) and a maxTokens budget.
// Splits by lines, takes head and tail portions, inserts a separator.
func CompressHeadTail(text string, headRatio, tailRatio float64, maxTokens int) string {
	if types.EstimateTokens(text) <= maxTokens {
		return text
	}

	lines := strings.Split(text, "\n")
	totalLines := len(lines)
	if totalLines <= 2 {
		// Not enough lines to split meaningfully — just truncate chars
		return truncateToTokens(text, maxTokens)
	}

	// Estimate how many lines we can keep based on average tokens per line
	avgTokensPerLine := types.EstimateTokens(text) / totalLines
	if avgTokensPerLine == 0 {
		avgTokensPerLine = 1
	}

	// Reserve tokens for the separator line
	separatorReserve := 20
	availableTokens := maxTokens - separatorReserve
	if availableTokens <= 0 {
		availableTokens = maxTokens
	}

	budgetLines := availableTokens / avgTokensPerLine
	if budgetLines <= 0 {
		budgetLines = 1
	}
	if budgetLines >= totalLines {
		budgetLines = totalLines - 1
	}

	headLines := int(float64(budgetLines) * headRatio / (headRatio + tailRatio))
	tailLines := budgetLines - headLines

	if headLines <= 0 {
		headLines = 1
	}
	if tailLines <= 0 {
		tailLines = 0
	}
	if headLines+tailLines >= totalLines {
		return text
	}

	removedLines := totalLines - headLines - tailLines
	head := strings.Join(lines[:headLines], "\n")
	headTokens := types.EstimateTokens(head)
	tailTokens := 0
	var tail string
	if tailLines > 0 {
		tail = strings.Join(lines[totalLines-tailLines:], "\n")
		tailTokens = types.EstimateTokens(tail)
	}
	// Derive removed tokens without materializing the removed section.
	totalTokens := types.EstimateTokens(text)
	removedTokens := totalTokens - headTokens - tailTokens

	separator := fmt.Sprintf("\n... [compressed: removed %d lines, ~%d tokens] ...\n",
		removedLines, removedTokens)

	if tailLines > 0 {
		return head + separator + tail
	}
	return head + separator
}

// Compress applies the appropriate compressor based on content type.
func Compress(text string, contentType ContentType, maxTokens int, headRatio, tailRatio float64) string {
	if types.EstimateTokens(text) <= maxTokens {
		return text
	}

	switch contentType {
	case ContentTypeJSON:
		return CompressJSON(text, maxTokens)
	case ContentTypeLog:
		return CompressLog(text, maxTokens)
	case ContentTypeCode:
		return CompressCode(text, maxTokens)
	case ContentTypeStackTrace:
		return CompressStackTrace(text, maxTokens)
	default:
		return CompressHeadTail(text, headRatio, tailRatio, maxTokens)
	}
}

// truncateToTokens truncates text to approximately fit within a token budget.
func truncateToTokens(text string, maxTokens int) string {
	// Approximate: 4 chars per token for ASCII
	maxChars := maxTokens * 4
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars] + "... [truncated]"
}
