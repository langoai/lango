package memory

import (
	"unicode/utf8"

	"github.com/langowarny/lango/internal/session"
)

const (
	// asciiCharsPerToken is the approximate number of ASCII/Latin characters per token.
	asciiCharsPerToken = 4
	// cjkRunesPerToken is the approximate number of CJK runes per token.
	cjkRunesPerToken = 2
	// perMessageOverhead is the token overhead per message for role/formatting.
	perMessageOverhead = 4
)

// EstimateTokens returns a character-based token count approximation.
// ASCII/Latin characters are counted as 1 token per 4 characters.
// CJK characters (Chinese, Japanese, Korean) are counted as 1 token per 2 characters.
func EstimateTokens(text string) int {
	var asciiCount, cjkCount int
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		if isCJK(r) {
			cjkCount++
		} else {
			asciiCount++
		}
		i += size
	}

	tokens := asciiCount / asciiCharsPerToken
	tokens += cjkCount / cjkRunesPerToken
	return tokens
}

// CountMessageTokens returns the estimated token count for a single message.
func CountMessageTokens(msg session.Message) int {
	tokens := perMessageOverhead
	tokens += EstimateTokens(msg.Content)
	for _, tc := range msg.ToolCalls {
		tokens += EstimateTokens(tc.ID)
		tokens += EstimateTokens(tc.Name)
		tokens += EstimateTokens(tc.Input)
		tokens += EstimateTokens(tc.Output)
	}
	return tokens
}

// CountMessagesTokens returns the total estimated token count for a batch of messages.
func CountMessagesTokens(msgs []session.Message) int {
	var total int
	for _, msg := range msgs {
		total += CountMessageTokens(msg)
	}
	return total
}

// isCJK returns true if the rune falls within CJK Unicode ranges.
func isCJK(r rune) bool {
	// CJK Unified Ideographs
	if r >= 0x4E00 && r <= 0x9FFF {
		return true
	}
	// CJK Unified Ideographs Extension A
	if r >= 0x3400 && r <= 0x4DBF {
		return true
	}
	// Korean Hangul Syllables
	if r >= 0xAC00 && r <= 0xD7AF {
		return true
	}
	return false
}
