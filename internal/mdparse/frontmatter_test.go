package mdparse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitFrontmatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		wantFM   string
		wantBody string
		wantErr  bool
	}{
		{
			give:     "---\ntitle: hello\n---\nbody text",
			wantFM:   "title: hello\n",
			wantBody: "body text",
		},
		{
			give:     "---\nkey: value\ntags:\n  - a\n  - b\n---\n\n# Heading\n\nParagraph here.",
			wantFM:   "key: value\ntags:\n  - a\n  - b\n",
			wantBody: "# Heading\n\nParagraph here.",
		},
		{
			give:     "---\n---\nbody only",
			wantFM:   "",
			wantBody: "body only",
		},
		{
			give:     "---\nfoo: bar\n---",
			wantFM:   "foo: bar\n",
			wantBody: "",
		},
		{
			give:     "---\n---",
			wantFM:   "",
			wantBody: "",
		},
		{
			give:    "",
			wantErr: true,
		},
		{
			give:    "no frontmatter here",
			wantErr: true,
		},
		{
			give:    "---\nunclosed frontmatter without closing delimiter",
			wantErr: true,
		},
		{
			give:     "  \n\n---\ntitle: trimmed\n---\nbody",
			wantFM:   "title: trimmed\n",
			wantBody: "body",
		},
		{
			give:     "---\r\ntitle: crlf\r\n---\r\nbody with crlf",
			wantFM:   "title: crlf\r\n",
			wantBody: "body with crlf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			fm, body, err := SplitFrontmatter([]byte(tt.give))

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantFM, string(fm))
			assert.Equal(t, tt.wantBody, body)
		})
	}
}

func TestSplitFrontmatter_NilInput(t *testing.T) {
	t.Parallel()

	_, _, err := SplitFrontmatter(nil)
	require.Error(t, err)
}
