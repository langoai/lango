// Package doctor implements the lango doctor command.
package doctor

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cliboot"
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

Checks performed (27 total):

  Core Configuration:
    - Configuration Profile   Profile validity and file accessibility
    - AI Providers            Provider configuration and API keys
    - API Key Security        Env-var best practices, secret exposure
    - Channel Tokens          Telegram, Discord, Slack token validation
    - Session Database        Database accessibility and integrity
    - Server Port             Port availability and binding

  Security:
    - Security Configuration  Signer, interceptor, encryption settings
    - Companion Connectivity  WebSocket gateway reachability

  Context Engineering:
    - Context Engineering     Context profile, budget allocation ratios
    - Retrieval Coordinator   Retrieval coordinator and auto-adjust settings

  Memory & Scanning:
    - Observational Memory    Memory configuration and storage
    - Output Scanning         Interceptor and scanning settings

  Embedding / RAG:
    - Embedding / RAG         Provider, model, and vector store setup

  Graph / Multi-Agent / A2A:
    - Graph Store             Triple store configuration
    - Multi-Agent             Orchestration and sub-agent settings
    - A2A Protocol            Agent-to-agent protocol connectivity

  Tool Hooks & Agent Management:
    - Tool Hooks              Middleware chain and hook configuration
    - Agent Registry          Registered agents and routing
    - Proactive Librarian     Librarian agent settings
    - Approval System         Tool approval workflow configuration

  Execution:
    - RunLedger               Ledger configuration invariants
    - Bootstrap Timing        Phase timing baseline comparison
    - RunLedger Workspace Isolation  Worktree health and stale detection

  Economy / Contract / Observability:
    - Economy Layer           Token economy and budget settings
    - Smart Contracts         Contract validation and deployment
    - Observability           Tracing, metrics, and logging

  P2P Workspace:
    - P2P Workspaces          Workspace isolation and connectivity

Use --fix to attempt automatic repair of fixable issues.
Use --json for machine-readable output.

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
	boot, err := bootstrap.Run(bootstrap.Options{Version: cliboot.Version})
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
		if bootAware, ok := check.(checks.BootstrapAwareCheck); ok {
			result = bootAware.RunWithBootstrap(ctx, cfg, boot)
		}

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
