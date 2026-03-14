package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewLoggingForm creates the Logging configuration form.
func NewLoggingForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Logging Configuration")

	form.AddField(&tuicore.Field{
		Key: "log_level", Label: "Log Level", Type: tuicore.InputSelect,
		Value:       cfg.Logging.Level,
		Options:     []string{"debug", "info", "warn", "error"},
		Description: "Application log verbosity (debug, info, warn, error)",
	})

	form.AddField(&tuicore.Field{
		Key: "log_format", Label: "Log Format", Type: tuicore.InputSelect,
		Value:       cfg.Logging.Format,
		Options:     []string{"console", "json"},
		Description: "Log output format (console for human-readable, json for structured)",
	})

	form.AddField(&tuicore.Field{
		Key: "log_output_path", Label: "Output Path", Type: tuicore.InputText,
		Value:       cfg.Logging.OutputPath,
		Placeholder: "stdout (leave empty for stdout)",
		Description: "File path for log output (empty = stdout)",
	})

	return &form
}

// NewGatekeeperForm creates the Gatekeeper (response sanitization) configuration form.
func NewGatekeeperForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Gatekeeper Configuration")

	form.AddField(&tuicore.Field{
		Key: "gk_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.Gatekeeper.Enabled, true),
		Description: "Enable response sanitization (strips internal markers, thought tags, etc.)",
	})

	form.AddField(&tuicore.Field{
		Key: "gk_strip_thought_tags", Label: "Strip Thought Tags", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.Gatekeeper.StripThoughtTags, true),
		Description: "Remove <thought>/<thinking> tags from LLM responses",
	})

	form.AddField(&tuicore.Field{
		Key: "gk_strip_internal_markers", Label: "Strip Internal Markers", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.Gatekeeper.StripInternalMarkers, true),
		Description: "Remove lines starting with [INTERNAL], [DEBUG], [SYSTEM], [OBSERVATION]",
	})

	form.AddField(&tuicore.Field{
		Key: "gk_strip_raw_json", Label: "Strip Raw JSON", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.Gatekeeper.StripRawJSON, true),
		Description: "Replace large raw JSON code blocks with a summary placeholder",
	})

	form.AddField(&tuicore.Field{
		Key: "gk_raw_json_threshold", Label: "Raw JSON Threshold", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Gatekeeper.RawJSONThreshold),
		Placeholder: "500 (character count)",
		Description: "Minimum characters for raw JSON replacement",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "gk_custom_patterns", Label: "Custom Patterns", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Gatekeeper.CustomPatterns, ","),
		Placeholder: "regex1,regex2 (comma-separated)",
		Description: "Additional regex patterns to strip from responses",
	})

	return &form
}

// NewOutputManagerForm creates the Output Manager configuration form.
func NewOutputManagerForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Output Manager Configuration")

	form.AddField(&tuicore.Field{
		Key: "om_mgr_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.Tools.OutputManager.Enabled, true),
		Description: "Enable token-based output compression for tool results",
	})

	form.AddField(&tuicore.Field{
		Key: "om_mgr_token_budget", Label: "Token Budget", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Tools.OutputManager.TokenBudget),
		Placeholder: "2000",
		Description: "Maximum token budget for tool output before compression",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_mgr_head_ratio", Label: "Head Ratio", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.2f", cfg.Tools.OutputManager.HeadRatio),
		Placeholder: "0.70 (0.0 to 1.0)",
		Description: "Ratio of head content to preserve during compression",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_mgr_tail_ratio", Label: "Tail Ratio", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.2f", cfg.Tools.OutputManager.TailRatio),
		Placeholder: "0.30 (0.0 to 1.0)",
		Description: "Ratio of tail content to preserve during compression",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
	})

	return &form
}
