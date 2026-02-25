// Package doctor implements the lango doctor command.
package doctor

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/doctor/checks"
	"github.com/langoai/lango/internal/cli/doctor/output"
	"github.com/langoai/lango/internal/config"
)

// Options holds the doctor command options.
type Options struct {
	Fix  bool
	JSON bool
}

// NewCommand creates the doctor command.
func NewCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose and fix Lango configuration issues",
		Long: `The doctor command checks your Lango configuration and environment
for common issues and can automatically fix some problems.

Checks performed:
  - Encrypted configuration profile validity
  - API key and provider configuration
  - Channel token validation
  - Session database accessibility
  - Server port availability
  - Security configuration (signer, interceptor, passphrase)
  - Companion connectivity (WebSocket gateway status)

See Also:
  lango settings - Interactive settings editor (TUI)
  lango config   - View/manage configuration profiles
  lango onboard  - Guided setup wizard`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Fix, "fix", false, "Attempt to automatically fix issues")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output results as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	// Load configuration from encrypted profile via bootstrap.
	var cfg *config.Config
	boot, err := bootstrap.Run(bootstrap.Options{})
	if err == nil {
		cfg = boot.Config
		defer boot.DBClient.Close()
	}

	// Get all checks
	allChecks := checks.AllChecks()
	results := make([]checks.Result, 0, len(allChecks))

	// Run checks
	for _, check := range allChecks {
		result := check.Run(ctx, cfg)

		// Try to fix if --fix is enabled and issue is fixable
		if opts.Fix && result.Fixable && result.Status == checks.StatusFail {
			result = check.Fix(ctx, cfg)
		}

		results = append(results, result)
	}

	summary := checks.NewSummary(results)

	// Output results
	if opts.JSON {
		renderer := &output.JSONRenderer{}
		jsonOutput, err := renderer.Render(summary)
		if err != nil {
			return fmt.Errorf("render JSON: %w", err)
		}
		fmt.Println(jsonOutput)
	} else {
		renderer := &output.TUIRenderer{}
		fmt.Print(renderer.RenderTitle())
		fmt.Println()

		for _, result := range results {
			fmt.Print(renderer.RenderResult(result))
		}

		fmt.Print(renderer.RenderSummary(summary))

		// Show fix hint if there are fixable issues
		hasFixable := false
		for _, result := range results {
			if result.Fixable && result.Status == checks.StatusFail {
				hasFixable = true
				break
			}
		}
		fmt.Print(renderer.RenderFixHint(hasFixable))
	}

	// Return error if there are failures
	if summary.HasErrors() {
		return fmt.Errorf("doctor found %d error(s)", summary.Failed)
	}

	return nil
}
