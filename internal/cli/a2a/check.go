package a2a

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "check <url>",
		Short: "Fetch and display a remote agent card",
		Long:  `Fetch the agent card from a remote A2A agent URL and display its contents.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(card)
			}

			fmt.Printf("Remote Agent Card\n")
			fmt.Printf("  Name:         %s\n", card.Name)
			fmt.Printf("  Description:  %s\n", card.Description)
			fmt.Printf("  URL:          %s\n", card.URL)
			if card.DID != "" {
				fmt.Printf("  DID:          %s\n", card.DID)
			}
			if len(card.Capabilities) > 0 {
				fmt.Printf("  Capabilities: %v\n", card.Capabilities)
			}
			fmt.Println()

			if len(card.Skills) > 0 {
				fmt.Printf("Skills (%d)\n", len(card.Skills))
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
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

			fmt.Println("No skills advertised.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
