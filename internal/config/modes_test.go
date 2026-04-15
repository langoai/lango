package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuiltInModes_ExposesThreeModes(t *testing.T) {
	modes := BuiltInModes()
	for _, name := range []string{"code-review", "research", "debug"} {
		if _, ok := modes[name]; !ok {
			t.Errorf("built-in mode %q missing", name)
		}
	}
}

func TestResolveModes_MergesUserDefinedWithBuiltIns(t *testing.T) {
	cfg := &Config{
		Modes: map[string]SessionMode{
			"custom": {Name: "custom", Tools: []string{"a", "b"}},
		},
	}
	resolved := cfg.ResolveModes()
	if _, ok := resolved["code-review"]; !ok {
		t.Fatal("built-in mode should remain after merge")
	}
	if _, ok := resolved["custom"]; !ok {
		t.Fatal("user-defined mode should be present")
	}
}

func TestResolveModes_UserOverridesBuiltIn(t *testing.T) {
	cfg := &Config{
		Modes: map[string]SessionMode{
			"code-review": {Name: "code-review", SystemHint: "custom hint"},
		},
	}
	resolved := cfg.ResolveModes()
	assert.Equal(t, "custom hint", resolved["code-review"].SystemHint)
}

func TestResolveModes_UserEntryMissingNameDefaultsToKey(t *testing.T) {
	cfg := &Config{
		Modes: map[string]SessionMode{
			"tagless": {Tools: []string{"x"}},
		},
	}
	resolved := cfg.ResolveModes()
	assert.Equal(t, "tagless", resolved["tagless"].Name)
}

func TestLookupMode_Found(t *testing.T) {
	cfg := &Config{}
	m, ok := cfg.LookupMode("research")
	assert.True(t, ok)
	assert.Equal(t, "research", m.Name)
}

func TestLookupMode_EmptyName(t *testing.T) {
	cfg := &Config{}
	_, ok := cfg.LookupMode("")
	assert.False(t, ok)
}

func TestLookupMode_NotFound(t *testing.T) {
	cfg := &Config{}
	_, ok := cfg.LookupMode("nonexistent")
	assert.False(t, ok)
}
