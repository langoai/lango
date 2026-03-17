package gatekeeper

import (
	"strings"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool { return &b }

func TestNewSanitizer_InvalidCustomPattern(t *testing.T) {
	cfg := config.GatekeeperConfig{
		CustomPatterns: []string{"[invalid"},
	}
	_, err := NewSanitizer(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compile custom pattern")
}

func TestSanitizer_Sanitize(t *testing.T) {
	tests := []struct {
		give string
		cfg  config.GatekeeperConfig
		want string
	}{
		{
			give: "Hello <thought>internal reasoning here</thought> world",
			cfg:  config.GatekeeperConfig{},
			want: "Hello  world",
		},
		{
			give: "Hello <thinking>deep analysis\nmultiline</thinking> world",
			cfg:  config.GatekeeperConfig{},
			want: "Hello  world",
		},
		{
			give: "Here is code:\n```\n<thought>this is inside a code block</thought>\n```\nDone",
			cfg:  config.GatekeeperConfig{},
			want: "Here is code:\n```\n<thought>this is inside a code block</thought>\n```\nDone",
		},
		{
			give: "[INTERNAL] secret debug info\nvisible line\n[DEBUG] more debug\n[SYSTEM] system note\n[OBSERVATION] obs",
			cfg:  config.GatekeeperConfig{},
			want: "visible line",
		},
		{
			give: "Before\n```json\n" + strings.Repeat("x", 600) + "\n```\nAfter",
			cfg:  config.GatekeeperConfig{},
			want: "Before\n[Large data block omitted]\nAfter",
		},
		{
			give: "Before\n```json\n{\"small\": true}\n```\nAfter",
			cfg:  config.GatekeeperConfig{},
			want: "Before\n```json\n{\"small\": true}\n```\nAfter",
		},
		{
			give: "Hello SECRET world",
			cfg: config.GatekeeperConfig{
				CustomPatterns: []string{`SECRET\s*`},
			},
			want: "Hello world",
		},
		{
			give: "line1\n\n\n\n\nline2",
			cfg:  config.GatekeeperConfig{},
			want: "line1\n\nline2",
		},
		{
			give: "Hello world, nothing to sanitize",
			cfg:  config.GatekeeperConfig{},
			want: "Hello world, nothing to sanitize",
		},
		{
			give: "Hello <thought>secret</thought> world",
			cfg: config.GatekeeperConfig{
				Enabled: boolPtr(false),
			},
			want: "Hello <thought>secret</thought> world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give[:min(len(tt.give), 40)], func(t *testing.T) {
			san, err := NewSanitizer(tt.cfg)
			require.NoError(t, err)
			got := san.Sanitize(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizer_Enabled(t *testing.T) {
	tests := []struct {
		give *bool
		want bool
	}{
		{give: nil, want: true},
		{give: boolPtr(true), want: true},
		{give: boolPtr(false), want: false},
	}
	for _, tt := range tests {
		san, err := NewSanitizer(config.GatekeeperConfig{Enabled: tt.give})
		require.NoError(t, err)
		assert.Equal(t, tt.want, san.Enabled())
	}
}

func TestSanitizer_ChunkSanitization(t *testing.T) {
	tests := []struct {
		give     string
		wantEmpty bool
		want     string
	}{
		{
			give: "<thought>entire chunk is internal</thought>",
			wantEmpty: true,
		},
		{
			give: "visible text <thought>hidden</thought>",
			want: "visible text",
		},
		{
			give: "[INTERNAL] debug line only",
			wantEmpty: true,
		},
		{
			give: "clean chunk with no internal content",
			want: "clean chunk with no internal content",
		},
	}

	san, err := NewSanitizer(config.GatekeeperConfig{})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.give[:min(len(tt.give), 40)], func(t *testing.T) {
			got := san.Sanitize(tt.give)
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSanitizer_ChannelIntegrationPath(t *testing.T) {
	// Simulates the runAgent() → Sanitize() integration path.
	san, err := NewSanitizer(config.GatekeeperConfig{})
	require.NoError(t, err)

	// Response with mixed internal and user-visible content.
	response := "Here are the results:\n<thought>I need to summarize this carefully</thought>\n[INTERNAL] raw debug\n\nThe search found 3 matching files."
	got := san.Sanitize(response)

	assert.NotContains(t, got, "<thought>")
	assert.NotContains(t, got, "[INTERNAL]")
	assert.Contains(t, got, "Here are the results:")
	assert.Contains(t, got, "The search found 3 matching files.")
}

func TestSanitizer_RawJSONThreshold(t *testing.T) {
	input := "Before\n```json\n" + strings.Repeat("a", 100) + "\n```\nAfter"

	san, err := NewSanitizer(config.GatekeeperConfig{
		RawJSONThreshold: 50,
	})
	require.NoError(t, err)
	got := san.Sanitize(input)
	assert.Equal(t, "Before\n[Large data block omitted]\nAfter", got)

	san2, err := NewSanitizer(config.GatekeeperConfig{
		RawJSONThreshold: 200,
	})
	require.NoError(t, err)
	got2 := san2.Sanitize(input)
	assert.Contains(t, got2, strings.Repeat("a", 100))
}
