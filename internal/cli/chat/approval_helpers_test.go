package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStringer string

func (s testStringer) String() string { return string(s) }

func TestSortedParamKeys(t *testing.T) {
	assert.Nil(t, sortedParamKeys(nil))
	assert.Nil(t, sortedParamKeys(map[string]any{}))
	assert.Equal(t, []string{"alpha", "beta", "zeta"}, sortedParamKeys(map[string]any{
		"zeta":  1,
		"alpha": 2,
		"beta":  3,
	}))
}

func TestSingleLineValue(t *testing.T) {
	assert.Equal(t, "", singleLineValue(""))
	assert.Equal(t, "line one line two", singleLineValue("line one\nline two"))
	assert.Equal(t, "trimmed value", singleLineValue("  trimmed \t value  "))
}

func TestFormatParamValue(t *testing.T) {
	assert.Equal(t, "null", formatParamValue(nil))
	assert.Equal(t, "line one line two", formatParamValue("line one\nline two"))
	assert.Equal(t, "stringer value", formatParamValue(testStringer("stringer\nvalue")))
	assert.Equal(t, `{"alpha":1,"zeta":2}`, formatParamValue(map[string]any{
		"zeta":  2,
		"alpha": 1,
	}))
	assert.Equal(t, "42", formatParamValue(42))
}

func TestCompactRequestID(t *testing.T) {
	assert.Equal(t, "", compactRequestID(""))
	assert.Equal(t, "apr-1", compactRequestID("apr-1"))
	assert.Equal(t, "req-1234…7890", compactRequestID("req-12345678901234567890"))
	assert.Equal(t, "req-1234 7890", compactRequestID("req-\x1b[31m1234\n7890\x1b[0m"))
}
