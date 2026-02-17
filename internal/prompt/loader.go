package prompt

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// sectionFileInfo maps a known filename to its section metadata.
type sectionFileInfo struct {
	ID       SectionID
	Priority int
	Title    string
}

// sectionFiles maps filename → section metadata.
var sectionFiles = map[string]sectionFileInfo{
	"AGENTS.md":             {SectionIdentity, 100, ""},
	"SAFETY.md":             {SectionSafety, 200, "Safety Guidelines"},
	"CONVERSATION_RULES.md": {SectionConversationRules, 300, "Conversation Rules"},
	"TOOL_USAGE.md":         {SectionToolUsage, 400, "Tool Usage Guidelines"},
}

// agentSectionFiles maps known per-agent filenames to section metadata.
// IDENTITY.md uses SectionAgentIdentity (priority 150) to sit between
// the global identity (100) and safety (200).
var agentSectionFiles = map[string]sectionFileInfo{
	"IDENTITY.md":           {SectionAgentIdentity, 150, ""},
	"SAFETY.md":             {SectionSafety, 200, "Safety Guidelines"},
	"CONVERSATION_RULES.md": {SectionConversationRules, 300, "Conversation Rules"},
}

// LoadAgentFromDir overlays per-agent prompt overrides on top of a shared
// base builder. The directory should be <promptsDir>/agents/<agentName>/.
// If the directory does not exist, the base builder is returned unmodified.
func LoadAgentFromDir(base *Builder, dir string, logger *zap.SugaredLogger) *Builder {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// Directory does not exist — return base as-is.
		return base
	}

	b := base.Clone()

	customPriority := 900
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			if logger != nil {
				logger.Warnw("agent prompt file read error", "file", entry.Name(), "error", err)
			}
			continue
		}

		text := strings.TrimSpace(string(content))
		if text == "" {
			continue
		}

		if info, ok := agentSectionFiles[entry.Name()]; ok {
			b.Add(NewStaticSection(info.ID, info.Priority, info.Title, text))
			if logger != nil {
				logger.Infow("loaded agent prompt override", "file", entry.Name(), "section", info.ID)
			}
		} else {
			// Unknown .md → custom section
			name := strings.TrimSuffix(entry.Name(), ".md")
			id := SectionID("custom_" + strings.ToLower(name))
			title := strings.ReplaceAll(name, "_", " ")
			b.Add(NewStaticSection(id, customPriority, title, text))
			customPriority++
			if logger != nil {
				logger.Infow("loaded agent custom prompt section", "file", entry.Name(), "section", id)
			}
		}
	}

	return b
}

// LoadFromDir loads .md files from the given directory and overrides
// matching default sections. Unknown .md files are added as custom
// sections with priority 900+.
func LoadFromDir(dir string, logger *zap.SugaredLogger) *Builder {
	b := DefaultBuilder()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if logger != nil {
			logger.Warnw("prompts directory not readable, using defaults", "dir", dir, "error", err)
		}
		return b
	}

	customPriority := 900
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			if logger != nil {
				logger.Warnw("prompt file read error", "file", entry.Name(), "error", err)
			}
			continue
		}

		text := strings.TrimSpace(string(content))
		if text == "" {
			continue
		}

		if info, ok := sectionFiles[entry.Name()]; ok {
			b.Add(NewStaticSection(info.ID, info.Priority, info.Title, text))
			if logger != nil {
				logger.Infow("loaded prompt section from file", "file", entry.Name(), "section", info.ID)
			}
		} else {
			// Unknown .md → custom section
			name := strings.TrimSuffix(entry.Name(), ".md")
			id := SectionID("custom_" + strings.ToLower(name))
			title := strings.ReplaceAll(name, "_", " ")
			b.Add(NewStaticSection(id, customPriority, title, text))
			customPriority++
			if logger != nil {
				logger.Infow("loaded custom prompt section", "file", entry.Name(), "section", id)
			}
		}
	}

	return b
}
