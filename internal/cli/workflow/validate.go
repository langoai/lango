package workflow

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/workflow"
)

func newValidateCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "validate <file.flow.yaml>",
		Short:         "Validate a workflow YAML file without executing",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			filePath := args[0]

			w, err := workflow.ParseFile(filePath)
			if err != nil {
				if output == "json" {
					return printJSON(cmd.OutOrStdout(), map[string]interface{}{
						"valid": false,
						"file":  filePath,
						"error": err.Error(),
					})
				}
				return fmt.Errorf("validate %q: %w", filePath, err)
			}

			type validateOutput struct {
				Valid    bool   `json:"valid"`
				File     string `json:"file"`
				Name     string `json:"name"`
				Steps    int    `json:"steps"`
				Schedule string `json:"schedule,omitempty"`
			}

			out := validateOutput{
				Valid:    true,
				File:     filePath,
				Name:     w.Name,
				Steps:    len(w.Steps),
				Schedule: w.Schedule,
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Workflow %q is valid.\n", filePath)
			fmt.Fprintf(cmd.OutOrStdout(), "  Name:     %s\n", out.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Steps:    %d\n", out.Steps)
			if out.Schedule != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  Schedule: %s\n", out.Schedule)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	return cmd
}
