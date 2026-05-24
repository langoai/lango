package a2a

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "check <url>",
		Short:         "Fetch and display a remote agent card",
		Long:          `Fetch the agent card from a remote A2A agent URL and display its contents.`,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			url := args[0]

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("fetch agent card: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("fetch agent card: HTTP %d", resp.StatusCode)
			}

			body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB limit
			if err != nil {
				return fmt.Errorf("read response: %w", err)
			}

			type skill struct {
				ID   string   `json:"id"`
				Name string   `json:"name"`
				Tags []string `json:"tags,omitempty"`
			}

			type remoteCard struct {
				Name         string   `json:"name"`
				Description  string   `json:"description"`
				URL          string   `json:"url"`
				Skills       []skill  `json:"skills,omitempty"`
				DID          string   `json:"did,omitempty"`
				Capabilities []string `json:"capabilities,omitempty"`
			}

			var card remoteCard
			if err := json.Unmarshal(body, &card); err != nil {
				return fmt.Errorf("parse agent card: %w", err)
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), card)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Remote Agent Card\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Name:         %s\n", card.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Description:  %s\n", card.Description)
			fmt.Fprintf(cmd.OutOrStdout(), "  URL:          %s\n", card.URL)
			if card.DID != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  DID:          %s\n", card.DID)
			}
			if len(card.Capabilities) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Capabilities: %v\n", card.Capabilities)
			}
			fmt.Fprintln(cmd.OutOrStdout())

			if len(card.Skills) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Skills (%d)\n", len(card.Skills))
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "  ID\tNAME\tTAGS")
				for _, s := range card.Skills {
					tags := "-"
					if len(s.Tags) > 0 {
						tags = fmt.Sprintf("%v", s.Tags)
					}
					fmt.Fprintf(w, "  %s\t%s\t%s\n", s.ID, s.Name, tags)
				}
				return w.Flush()
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "No skills advertised.")
			return err
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
