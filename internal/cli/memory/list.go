package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/session"
	"github.com/spf13/cobra"
)

func newListCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		sessionKey string
		memType    string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List observations and reflections for a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, memStore, cleanup, err := initMemoryStore(cfg)
			if err != nil {
				return err
			}
			_ = store
			defer cleanup()

			ctx := context.Background()

			type entry struct {
				ID        string    `json:"id"`
				Type      string    `json:"type"`
				Tokens    int       `json:"tokens"`
				CreatedAt time.Time `json:"created_at"`
				Content   string    `json:"content"`
			}

			var entries []entry

			if memType == "" || memType == "observations" {
				obs, err := memStore.ListObservations(ctx, sessionKey)
				if err != nil {
					return fmt.Errorf("list observations: %w", err)
				}
				for _, o := range obs {
					entries = append(entries, entry{
						ID:        o.ID.String(),
						Type:      "observation",
						Tokens:    o.TokenCount,
						CreatedAt: o.CreatedAt,
						Content:   o.Content,
					})
				}
			}

			if memType == "" || memType == "reflections" {
				refs, err := memStore.ListReflections(ctx, sessionKey)
				if err != nil {
					return fmt.Errorf("list reflections: %w", err)
				}
				for _, r := range refs {
					entries = append(entries, entry{
						ID:        r.ID.String(),
						Type:      "reflection",
						Tokens:    r.TokenCount,
						CreatedAt: r.CreatedAt,
						Content:   r.Content,
					})
				}
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(entries)
			}

			if len(entries) == 0 {
				fmt.Println("No entries found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTYPE\tTOKENS\tCREATED\tCONTENT")
			for _, e := range entries {
				content := e.Content
				if len(content) > 60 {
					content = content[:57] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
					e.ID[:8], e.Type, e.Tokens,
					e.CreatedAt.Format("2006-01-02 15:04"),
					content,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&sessionKey, "session", "", "Session key (required)")
	cmd.Flags().StringVar(&memType, "type", "", "Filter by type: observations, reflections")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("session")

	return cmd
}

func initMemoryStore(cfg *config.Config) (*session.EntStore, *memory.Store, func(), error) {
	store, err := session.NewEntStore(cfg.Session.DatabasePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open session store: %w", err)
	}

	logger := zap.NewNop().Sugar()
	memStore := memory.NewStore(store.Client(), logger)

	cleanup := func() {
		store.Close()
	}

	return store, memStore, cleanup, nil
}
