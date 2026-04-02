package agentregistry

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/langoai/lango/internal/mdparse"
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

	body := def.Instruction
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}

	result, err := mdparse.RenderFrontmatter(&m, body)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}
	return result, nil
}

// splitFrontmatter delegates to mdparse.SplitFrontmatter.
var splitFrontmatter = mdparse.SplitFrontmatter
