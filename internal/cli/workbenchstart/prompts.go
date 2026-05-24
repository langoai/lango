package workbenchstart

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var defaultPrompts = []string{
	"Summarize this repository",
	"Explain the current project structure",
	"Review recent changes",
}

type workspaceContext struct {
	rootName       string
	hasGit         bool
	hasGoMod       bool
	branch         string
	dirty          bool
	changedTargets []string
}

func DefaultPrompts() []string {
	return append([]string(nil), defaultPrompts...)
}

func DefaultPrompt(workDir string) string {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		prompts := DefaultPrompts()
		if len(prompts) == 0 {
			return ""
		}
		return prompts[0]
	}
	ctx := inspectWorkspaceContext(workDir)
	prompts := buildPromptsFromContext(ctx)
	if len(prompts) == 0 {
		return ""
	}
	if ctx.hasGit && ctx.dirty && len(prompts) >= 3 {
		return prompts[2]
	}
	return prompts[0]
}

func PostTurnDefaultPrompt(workDir string) string {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		prompts := DefaultPrompts()
		if len(prompts) >= 2 {
			return prompts[1]
		}
		if len(prompts) == 1 {
			return prompts[0]
		}
		return ""
	}

	ctx := inspectWorkspaceContext(workDir)
	prompts := buildPromptsFromContext(ctx)
	if len(prompts) == 0 {
		return ""
	}
	if ctx.rootName == "" {
		if len(prompts) >= 2 {
			return prompts[1]
		}
		return prompts[0]
	}
	if len(prompts) >= 3 {
		return prompts[2]
	}
	return prompts[0]
}

func RecoveryDefaultPrompt(workDir string) string {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		return preferredPrompt(DefaultPrompts(), 2)
	}

	ctx := inspectWorkspaceContext(workDir)
	prompts := buildPromptsFromContext(ctx)
	return preferredPrompt(prompts, 2)
}

func BuildPrompts(workDir string) []string {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		return DefaultPrompts()
	}
	ctx := inspectWorkspaceContext(workDir)
	return buildPromptsFromContext(ctx)
}

func buildPromptsFromContext(ctx workspaceContext) []string {
	if ctx.rootName == "" {
		return DefaultPrompts()
	}

	prompts := []string{
		"Summarize the " + ctx.rootName + " repository and its current purpose",
		"Explain the current project structure in " + ctx.rootName,
		"Review the likely active workstream in " + ctx.rootName + " and suggest the best next change",
	}
	if ctx.hasGoMod {
		prompts[1] = "Explain the Go package layout in " + ctx.rootName + " and where to start editing"
	}
	if !ctx.hasGit {
		prompts[2] = "Review the most important files in " + ctx.rootName + " and suggest the best next change"
	} else if ctx.branch != "" && ctx.dirty {
		prompts[2] = "Review the uncommitted changes on branch " + ctx.branch + " in " + ctx.rootName
		if focus := summarizeChangedTargets(ctx.changedTargets); focus != "" {
			prompts[2] += ", especially " + focus
		}
		prompts[2] += ", and suggest the best next edit"
	} else if ctx.branch != "" {
		prompts[2] = "Review the current state of branch " + ctx.branch + " in " + ctx.rootName + " and suggest the best next change"
	}
	return prompts
}

func inspectWorkspaceContext(workDir string) workspaceContext {
	resolved, err := filepath.Abs(workDir)
	if err != nil {
		return workspaceContext{}
	}
	resolved = filepath.Clean(resolved)

	var gitRoot string
	var goModRoot string
	for cur := resolved; ; cur = filepath.Dir(cur) {
		if gitRoot == "" && hasPath(filepath.Join(cur, ".git")) {
			gitRoot = cur
		}
		if goModRoot == "" && hasPath(filepath.Join(cur, "go.mod")) {
			goModRoot = cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
	}

	root := gitRoot
	if root == "" {
		root = goModRoot
	}
	if root == "" {
		return workspaceContext{}
	}
	if gitRoot == "" && goModRoot != "" {
		gitRoot = goModRoot
	}

	ctx := workspaceContext{
		rootName: filepath.Base(root),
		hasGit:   gitRoot != "",
		hasGoMod: goModRoot != "",
	}
	if ctx.hasGit {
		ctx.branch, ctx.dirty, ctx.changedTargets = inspectGitPromptSignals(gitRoot)
	}
	return ctx
}

func inspectGitPromptSignals(repoRoot string) (string, bool, []string) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", false, nil
	}

	branch, err := gitOutput(repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", false, nil
	}
	status, err := gitOutputRaw(repoRoot, "status", "--short")
	if err != nil {
		return branch, false, nil
	}
	dirty := strings.TrimSpace(status) != ""
	return branch, dirty, parseGitStatusTargets(status)
}

func gitOutput(repoRoot string, args ...string) (string, error) {
	output, err := gitOutputRaw(repoRoot, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func gitOutputRaw(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func parseGitStatusTargets(status string) []string {
	lines := strings.Split(status, "\n")
	seen := make(map[string]struct{}, len(lines))
	targets := make([]string, 0, len(lines))
	for _, raw := range lines {
		if len(raw) < 4 {
			continue
		}
		target := strings.TrimSpace(raw[3:])
		if idx := strings.LastIndex(target, " -> "); idx >= 0 {
			target = strings.TrimSpace(target[idx+4:])
		}
		if target == "" {
			continue
		}
		target = topLevelTarget(target)
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		targets = append(targets, target)
	}
	return targets
}

func topLevelTarget(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	if path == "" {
		return ""
	}
	if idx := strings.IndexByte(path, '/'); idx >= 0 {
		return path[:idx]
	}
	return path
}

func summarizeChangedTargets(targets []string) string {
	if len(targets) == 0 {
		return ""
	}
	switch len(targets) {
	case 1:
		return "`" + targets[0] + "`"
	case 2:
		return "`" + targets[0] + "` and `" + targets[1] + "`"
	default:
		return "`" + targets[0] + "`, `" + targets[1] + "`, and `" + targets[2] + "`"
	}
}

func preferredPrompt(prompts []string, idx int) string {
	if len(prompts) == 0 {
		return ""
	}
	if idx >= 0 && idx < len(prompts) {
		return prompts[idx]
	}
	return prompts[len(prompts)-1]
}

func hasPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
