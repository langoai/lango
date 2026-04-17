package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newDBMigrateCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "db-migrate",
		Short: "Legacy SQLCipher migration workflow (unsupported)",
		Long:  "Legacy SQLCipher database encryption workflows are no longer supported by this runtime.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()
			dbPath := resolveDBPath(boot.Config.Session.DatabasePath)
			_ = force
			return fmt.Errorf(
				"SQLCipher database encryption is no longer supported by this runtime. If %q is a legacy encrypted database, use an older build to export or decrypt it before upgrading",
				dbPath,
			)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

func newDBDecryptCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "db-decrypt",
		Short: "Legacy SQLCipher decrypt workflow (unsupported)",
		Long:  "Legacy SQLCipher database decryption workflows are no longer supported by this runtime.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()
			dbPath := resolveDBPath(boot.Config.Session.DatabasePath)
			_ = force
			return fmt.Errorf(
				"SQLCipher database decryption is no longer supported by this runtime. If %q is a legacy encrypted database, use an older build to export it before upgrading",
				dbPath,
			)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

// resolveDBPath expands tilde in a database path.
func resolveDBPath(dbPath string) string {
	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return dbPath
		}
		return filepath.Join(home, dbPath[2:])
	}
	return dbPath
}
