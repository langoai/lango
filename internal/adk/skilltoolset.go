// Package adk — Skill Toolset adapter
//
// Bridges Lango configuration to ADK's tool/skilltoolset. Reads markdown-defined
// skills from a directory and exposes them to the agent via three tools:
//
//   - load_skill            — read SKILL.md for a named skill
//   - list_skill_resources  — list references/, assets/, scripts/ under a skill
//   - load_skill_resource   — read one file inside a skill directory
//
// Each skill subdirectory MUST contain a SKILL.md with YAML frontmatter
// (name, description) per the agentskills.io specification. Skill directories
// without valid frontmatter are silently ignored by the upstream source.
package adk

import (
	"context"
	"errors"
	"fmt"
	"os"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/skilltoolset"
	"google.golang.org/adk/tool/skilltoolset/skill"
)

// ErrSkillsDirNotExist indicates the configured skill source directory does not exist.
var ErrSkillsDirNotExist = errors.New("skills directory does not exist")

// BuildSkillToolset constructs a SkillToolset from the given directory.
// Returns (nil, nil) if dir is empty (skill toolset disabled by config).
// Returns (nil, ErrSkillsDirNotExist) if dir is set but missing.
//
// The returned toolset implements tool.Toolset and can be passed to
// llmagent.Config.Toolsets via the WithToolsets agent option.
func BuildSkillToolset(ctx context.Context, dir string) (tool.Toolset, error) {
	if dir == "" {
		return nil, nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrSkillsDirNotExist, dir)
		}
		return nil, fmt.Errorf("stat skills dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("skills dir %q is not a directory", dir)
	}

	source := skill.NewFileSystemSource(os.DirFS(dir))
	ts, err := skilltoolset.New(ctx, skilltoolset.Config{Source: source})
	if err != nil {
		return nil, fmt.Errorf("create skill toolset: %w", err)
	}
	return ts, nil
}
