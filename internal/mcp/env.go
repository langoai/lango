package mcp

import (
	"os"
	"regexp"
	"strings"
)

var envVarWithDefaultRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

// ExpandEnv replaces ${VAR} and ${VAR:-default} patterns with environment variable values.
// If the variable is not set and no default is provided, the original pattern is kept.
func ExpandEnv(s string) string {
	return envVarWithDefaultRegex.ReplaceAllStringFunc(s, func(match string) string {
		inner := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")

		// Check for default value: ${VAR:-default}
		varName, defaultVal, hasDefault := strings.Cut(inner, ":-")

		if val := os.Getenv(varName); val != "" {
			return val
		}
		if hasDefault {
			return defaultVal
		}
		return match
	})
}

// ExpandEnvMap applies ExpandEnv to all values in a map, returning a new map.
func ExpandEnvMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = ExpandEnv(v)
	}
	return out
}

// BuildEnvSlice converts a map of env vars to a slice of "KEY=VALUE" strings
// suitable for os/exec.Cmd.Env, inheriting the current process environment.
func BuildEnvSlice(extra map[string]string) []string {
	if len(extra) == 0 {
		return nil
	}
	env := os.Environ()
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	return env
}
