package skill

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"text/template"

	"go.uber.org/zap"
)

var _dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`rm\s+-rf\s+/`),
	regexp.MustCompile(`:\(\)\s*\{.*\|.*&.*\};\s*:`),
	regexp.MustCompile(`(curl|wget).*\|\s*(sh|bash)`),
	regexp.MustCompile(`>\s*/dev/sd`),
	regexp.MustCompile(`mkfs\.`),
	regexp.MustCompile(`dd\s+if=`),
}

// Executor safely executes skills.
type Executor struct {
	logger *zap.SugaredLogger
}

// NewExecutor creates a new skill executor.
func NewExecutor(logger *zap.SugaredLogger) *Executor {
	return &Executor{logger: logger}
}

// Execute runs a skill with the given parameters.
func (e *Executor) Execute(ctx context.Context, skill SkillEntry, params map[string]interface{}) (interface{}, error) {
	switch skill.Type {
	case "composite":
		return e.executeComposite(ctx, skill)
	case "script":
		return e.executeScript(ctx, skill)
	case "template":
		return e.executeTemplate(skill, params)
	default:
		return nil, fmt.Errorf("unknown skill type: %s", skill.Type)
	}
}

// ValidateScript checks a script for dangerous patterns.
func (e *Executor) ValidateScript(script string) error {
	for _, pattern := range _dangerousPatterns {
		if pattern.MatchString(script) {
			return fmt.Errorf("script contains dangerous pattern: %s", pattern.String())
		}
	}
	return nil
}

func (e *Executor) executeComposite(_ context.Context, skill SkillEntry) (interface{}, error) {
	stepsRaw, ok := skill.Definition["steps"]
	if !ok {
		return nil, fmt.Errorf("composite skill %q missing 'steps' in definition", skill.Name)
	}

	steps, ok := stepsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("composite skill %q: 'steps' must be an array", skill.Name)
	}

	plan := make([]map[string]interface{}, 0, len(steps))
	for i, stepRaw := range steps {
		step, ok := stepRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("composite skill %q: step %d is not an object", skill.Name, i)
		}

		toolName, _ := step["tool"].(string)
		stepParams, _ := step["params"].(map[string]interface{})

		plan = append(plan, map[string]interface{}{
			"step":   i + 1,
			"tool":   toolName,
			"params": stepParams,
		})
	}

	return map[string]interface{}{
		"skill": skill.Name,
		"type":  "composite",
		"plan":  plan,
	}, nil
}

func (e *Executor) executeScript(ctx context.Context, skill SkillEntry) (interface{}, error) {
	scriptRaw, ok := skill.Definition["script"]
	if !ok {
		return nil, fmt.Errorf("script skill %q missing 'script' in definition", skill.Name)
	}

	script, ok := scriptRaw.(string)
	if !ok {
		return nil, fmt.Errorf("script skill %q: 'script' must be a string", skill.Name)
	}

	if err := e.ValidateScript(script); err != nil {
		return nil, fmt.Errorf("script skill %q: %w", skill.Name, err)
	}

	f, err := os.CreateTemp("", fmt.Sprintf("lango-skill-%s-*.sh", skill.Name))
	if err != nil {
		return nil, fmt.Errorf("create temp script: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write([]byte(script)); err != nil {
		f.Close()
		return nil, fmt.Errorf("write script: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close script: %w", err)
	}

	cmd := exec.CommandContext(ctx, "sh", f.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("execute script skill %q: %w (stderr: %s)", skill.Name, err, stderr.String())
	}

	return stdout.String(), nil
}

func (e *Executor) executeTemplate(skill SkillEntry, params map[string]interface{}) (interface{}, error) {
	tmplRaw, ok := skill.Definition["template"]
	if !ok {
		return nil, fmt.Errorf("template skill %q missing 'template' in definition", skill.Name)
	}

	tmplStr, ok := tmplRaw.(string)
	if !ok {
		return nil, fmt.Errorf("template skill %q: 'template' must be a string", skill.Name)
	}

	tmpl, err := template.New(skill.Name).Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("parse template skill %q: %w", skill.Name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, fmt.Errorf("execute template skill %q: %w", skill.Name, err)
	}

	return buf.String(), nil
}
