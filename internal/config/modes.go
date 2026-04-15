package config

// BuiltInModes returns the three shipping session modes: code-review, research,
// and debug. User-defined modes from Config.Modes merge on top of these; user
// entries with the same name replace the built-in.
func BuiltInModes() map[string]SessionMode {
	return map[string]SessionMode{
		"code-review": {
			Name: "code-review",
			Tools: []string{
				"@filesystem",
				"@exec",
				"builtin_search",
				"list_skills",
				"view_skill",
			},
			SystemHint: "You are in code-review mode. Focus on reading code, running tests, and reviewing diffs. Do not modify files without explicit user approval.",
		},
		"research": {
			Name: "research",
			Tools: []string{
				"@webfetch",
				"@websearch",
				"@filesystem",
				"builtin_search",
				"list_skills",
				"view_skill",
			},
			SystemHint: "You are in research mode. Prioritize gathering information from web and local documents. Avoid making system changes.",
		},
		"debug": {
			Name: "debug",
			Tools: []string{
				"@filesystem",
				"@exec",
				"builtin_search",
				"list_skills",
				"view_skill",
			},
			SystemHint: "You are in debug mode. Inspect logs, reproduce issues, and investigate root causes before proposing fixes.",
		},
	}
}

// ResolveModes returns the merged map of built-in and user-defined modes.
// User entries with the same name as a built-in replace the built-in entirely.
func (c *Config) ResolveModes() map[string]SessionMode {
	resolved := BuiltInModes()
	for name, mode := range c.Modes {
		if mode.Name == "" {
			mode.Name = name
		}
		resolved[name] = mode
	}
	return resolved
}

// LookupMode returns the SessionMode for the given name, or (zero, false) if
// not found. Built-in and user-defined modes are both searchable.
func (c *Config) LookupMode(name string) (SessionMode, bool) {
	if name == "" {
		return SessionMode{}, false
	}
	modes := c.ResolveModes()
	m, ok := modes[name]
	return m, ok
}
