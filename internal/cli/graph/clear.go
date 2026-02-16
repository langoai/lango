package graph

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/langowarny/lango/internal/config"
	"github.com/spf13/cobra"
)

func newClearCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all triples from the knowledge graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, err := initGraphStore(cfg)
			if err != nil {
				return err
			}
			defer store.Close()

			if !force {
				fmt.Println("This will delete all triples from the knowledge graph.")
				fmt.Print("Continue? [y/N] ")
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
					if answer != "y" && answer != "yes" {
						fmt.Println("Aborted.")
						return nil
					}
				}
			}

			if err := store.ClearAll(context.Background()); err != nil {
				return fmt.Errorf("clear graph: %w", err)
			}

			fmt.Println("Cleared all triples from the knowledge graph.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
