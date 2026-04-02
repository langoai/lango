package mdparse

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// RenderFrontmatter renders YAML frontmatter and a markdown body into the
// standard frontmatter format:
//
//	---
//	(YAML)
//	---
//
//	(body)
func RenderFrontmatter(meta interface{}, body string) ([]byte, error) {
	yamlBytes, err := yaml.Marshal(meta)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(body)
	return buf.Bytes(), nil
}
