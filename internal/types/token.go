package types

import "unicode/utf8"

const (
	// asciiCharsPerToken is the approximate number of ASCII/Latin characters per token.
	asciiCharsPerToken = 4
	// cjkRunesPerToken is the approximate number of CJK runes per token.
	cjkRunesPerToken = 2
)

// EstimateTokens returns a character-based token count approximation.
// ASCII/Latin characters are counted as 1 token per 4 characters.
// CJK characters (Chinese, Japanese, Korean) are counted as 1 token per 2 characters.
func EstimateTokens(text string) int {
	var asciiCount, cjkCount int
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		if IsCJK(r) {
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

// IsCJK returns true if the rune falls within CJK Unicode ranges.
func IsCJK(r rune) bool {
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
