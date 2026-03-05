package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandEnv(t *testing.T) {
	t.Setenv("TEST_TOKEN", "abc123")

	tests := []struct {
		give string
		want string
	}{
		{give: "plain", want: "plain"},
		{give: "${TEST_TOKEN}", want: "abc123"},
		{give: "Bearer ${TEST_TOKEN}", want: "Bearer abc123"},
		{give: "${UNSET_VAR}", want: "${UNSET_VAR}"},
		{give: "${UNSET_VAR:-fallback}", want: "fallback"},
		{give: "${TEST_TOKEN:-fallback}", want: "abc123"},
		{give: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, ExpandEnv(tt.give))
		})
	}
}

func TestExpandEnvMap(t *testing.T) {
	t.Setenv("MY_KEY", "secret")

	m := map[string]string{
		"KEY":   "${MY_KEY}",
		"PLAIN": "value",
	}

	got := ExpandEnvMap(m)
	assert.Equal(t, "secret", got["KEY"])
	assert.Equal(t, "value", got["PLAIN"])
}

func TestExpandEnvMap_Nil(t *testing.T) {
	assert.Nil(t, ExpandEnvMap(nil))
}

func TestBuildEnvSlice(t *testing.T) {
	got := BuildEnvSlice(map[string]string{"FOO": "bar"})
	assert.Contains(t, got, "FOO=bar")
	assert.True(t, len(got) > 1) // inherits os.Environ()
}

func TestBuildEnvSlice_Empty(t *testing.T) {
	assert.Nil(t, BuildEnvSlice(nil))
}
