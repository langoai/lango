package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticSection_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		title    string
		content  string
		wantText string
	}{
		{
			give:     "with title",
			title:    "Safety",
			content:  "Do not harm.",
			wantText: "## Safety\nDo not harm.",
		},
		{
			give:     "without title",
			title:    "",
			content:  "You are Lango.",
			wantText: "You are Lango.",
		},
		{
			give:     "empty content",
			title:    "Unused",
			content:  "",
			wantText: "",
		},
		{
			give:     "whitespace only content",
			title:    "Unused",
			content:  "   \n  ",
			wantText: "",
		},
		{
			give:     "content with leading/trailing whitespace",
			title:    "Rules",
			content:  "\n  Be concise.  \n",
			wantText: "## Rules\nBe concise.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			s := NewStaticSection("test", 100, tt.title, tt.content)
			assert.Equal(t, tt.wantText, s.Render())
		})
	}
}

func TestStaticSection_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	s := NewStaticSection(SectionIdentity, 100, "", "content")
	assert.Equal(t, SectionIdentity, s.ID())
	assert.Equal(t, 100, s.Priority())
}
