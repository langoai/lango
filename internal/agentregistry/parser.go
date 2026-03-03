package agentregistry

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseAgentMD parses an AGENT.md file (YAML frontmatter + markdown body).
func ParseAgentMD(content []byte) (*AgentDefinition, error) {
	fm, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, err
	}

	var def AgentDefinition
	if err := yaml.Unmarshal(fm, &def); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	if def.Name == "" {
		return nil, fmt.Errorf("agent name is required in frontmatter")
	}
	if def.Status == "" {
		def.Status = StatusActive
	}

	def.Instruction = body
	return &def, nil
}

// RenderAgentMD renders an AgentDefinition to AGENT.md format.
func RenderAgentMD(def *AgentDefinition) ([]byte, error) {
	status := def.Status
	if status == "" {
		status = StatusDraft
	}

	// Create a copy with the status set for marshaling.
	m := *def
	m.Status = status

	fmBytes, err := yaml.Marshal(&m)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(def.Instruction)
	if def.Instruction != "" && !strings.HasSuffix(def.Instruction, "\n") {
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// splitFrontmatter extracts YAML frontmatter and body from markdown content.
// Reuses the same pattern as skill/parser.go.
func splitFrontmatter(content []byte) (frontmatterBytes []byte, body string, err error) {
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
