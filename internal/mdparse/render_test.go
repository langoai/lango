package mdparse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderFrontmatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		giveMeta interface{}
		giveBody string
		want     string
	}{
		{
			give:     "simple struct",
			giveMeta: struct{ Name string `yaml:"name"` }{Name: "hello"},
			giveBody: "Body text.\n",
			want:     "---\nname: hello\n---\n\nBody text.\n",
		},
		{
			give:     "empty body",
			giveMeta: struct{ Name string `yaml:"name"` }{Name: "test"},
			giveBody: "",
			want:     "---\nname: test\n---\n\n",
		},
		{
			give: "map meta",
			giveMeta: map[string]string{
				"title": "doc",
			},
			giveBody: "# Heading\n\nParagraph.\n",
			want:     "---\ntitle: doc\n---\n\n# Heading\n\nParagraph.\n",
		},
		{
			give:     "body without trailing newline",
			giveMeta: struct{ Key string `yaml:"key"` }{Key: "val"},
			giveBody: "no newline",
			want:     "---\nkey: val\n---\n\nno newline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got, err := RenderFrontmatter(tt.giveMeta, tt.giveBody)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestRenderFrontmatter_Roundtrip(t *testing.T) {
	t.Parallel()

	type meta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}

	m := meta{Name: "roundtrip", Description: "test"}
	body := "Instruction body.\n"

	rendered, err := RenderFrontmatter(m, body)
	require.NoError(t, err)

	fm, parsedBody, err := SplitFrontmatter(rendered)
	require.NoError(t, err)
	assert.Contains(t, string(fm), "name: roundtrip")
	assert.Contains(t, string(fm), "description: test")
	assert.Equal(t, "Instruction body.", parsedBody)
}

func TestRenderFrontmatter_NilMeta(t *testing.T) {
	t.Parallel()

	// nil meta produces an empty YAML document ("null\n"), which is still valid output.
	got, err := RenderFrontmatter(nil, "body")
	require.NoError(t, err)
	assert.Contains(t, string(got), "---\n")
	assert.Contains(t, string(got), "body")
}
