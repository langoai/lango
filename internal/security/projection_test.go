package security

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestRedactedProjectionRedactsSecretsAndNormalizesWhitespace(t *testing.T) {
	got := RedactedProjection("  user@example.com\n token abcdef1234567890abcdef1234567890 123456  ", 0)

	assert.Equal(t, "[email] token [secret] [number]", got)
}

func TestRedactedProjectionTruncatesWithoutBreakingUTF8(t *testing.T) {
	limit := len("Hello ") + 1
	got := RedactedProjection("Hello 세계", limit)

	assert.True(t, utf8.ValidString(got), "projection must remain valid UTF-8")
	assert.False(t, strings.ContainsRune(got, utf8.RuneError), "projection must not contain replacement runes")
	assert.LessOrEqual(t, len(got), limit)
	assert.Equal(t, "Hello ", got)
}
