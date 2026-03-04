// Package mdparse provides shared markdown parsing utilities.
package mdparse

import (
	"fmt"
	"strings"
)

// SplitFrontmatter extracts YAML frontmatter and body from markdown content.
// The content must begin with a "---" delimiter line followed by YAML, then a
// closing "---" delimiter. Everything after the closing delimiter is returned
// as the body.
func SplitFrontmatter(content []byte) (frontmatterBytes []byte, body string, err error) {
	s := strings.TrimSpace(string(content))

	if !strings.HasPrefix(s, "---") {
		return nil, "", fmt.Errorf("missing frontmatter delimiter (---)")
	}

	rest := s[3:]
	rest = strings.TrimLeft(rest, "\r\n")
	idx := strings.Index(rest, "---")
	if idx < 0 {
		return nil, "", fmt.Errorf("missing closing frontmatter delimiter (---)")
	}

	fm := rest[:idx]
	body = strings.TrimSpace(rest[idx+3:])

	return []byte(fm), body, nil
}
