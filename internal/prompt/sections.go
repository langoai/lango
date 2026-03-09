package prompt

import "strings"

// StaticSection is a simple prompt section with fixed content.
type StaticSection struct {
	id       SectionID
	priority int
	title    string
	content  string
}

var _ PromptSection = (*StaticSection)(nil)

// NewStaticSection creates a new static prompt section.
func NewStaticSection(id SectionID, priority int, title, content string) *StaticSection {
	return &StaticSection{
		id:       id,
		priority: priority,
		title:    title,
		content:  content,
	}
}

func (s *StaticSection) ID() SectionID { return s.id }
func (s *StaticSection) Priority() int { return s.priority }

// Render returns the section content with an optional title header.
func (s *StaticSection) Render() string {
	content := strings.TrimSpace(s.content)
	if content == "" {
		return ""
	}
	if s.title == "" {
		return content
	}
	return "## " + s.title + "\n" + content
}
