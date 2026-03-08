package workflow

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/workflow"
)

func newValidateCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "validate <file.flow.yaml>",
		Short: "Validate a workflow YAML file without executing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			w, err := workflow.ParseFile(filePath)
			if err != nil {
				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(map[string]interface{}{
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("Workflow %q is valid.\n", filePath)
			fmt.Printf("  Name:     %s\n", out.Name)
			fmt.Printf("  Steps:    %d\n", out.Steps)
			if out.Schedule != "" {
				fmt.Printf("  Schedule: %s\n", out.Schedule)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
