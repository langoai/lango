package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/dbmigrate"
)

func newDBMigrateCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "db-migrate",
		Short: "Encrypt the application database with SQLCipher",
		Long:  "Converts the plaintext SQLite database to SQLCipher-encrypted format using the current passphrase.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			dbPath := resolveDBPath(boot.Config.Session.DatabasePath)
			if bootstrap.IsDBEncrypted(dbPath) {
				return fmt.Errorf("database is already encrypted")
			}

			if !force && !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal (use --force for non-interactive)")
			}

			if !force {
				ok, err := prompt.Confirm("This will encrypt your database. A backup will be created. Continue?")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Println("Aborted.")
					return nil
				}
			}

			pageSize := boot.Config.Security.DBEncryption.CipherPageSize
			if pageSize <= 0 {
				pageSize = 4096
			}

			pass, err := prompt.Passphrase("Enter passphrase for DB encryption: ")
			if err != nil {
				return fmt.Errorf("read passphrase: %w", err)
			}

			fmt.Println("Encrypting database...")
			if err := dbmigrate.MigrateToEncrypted(dbPath, pass, pageSize); err != nil {
				return fmt.Errorf("db encryption: %w", err)
			}

			fmt.Println("Database encrypted successfully.")
			fmt.Println("Set security.dbEncryption.enabled=true in your config to use the encrypted DB.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

func newDBDecryptCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "db-decrypt",
		Short: "Decrypt the application database back to plaintext",
		Long:  "Converts a SQLCipher-encrypted database back to a plaintext SQLite database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			dbPath := resolveDBPath(boot.Config.Session.DatabasePath)
			if !bootstrap.IsDBEncrypted(dbPath) {
				return fmt.Errorf("database is not encrypted")
			}

			if !force && !prompt.IsInteractive() {
				return fmt.Errorf("this command requires an interactive terminal (use --force for non-interactive)")
			}

			if !force {
				ok, err := prompt.Confirm("This will decrypt your database to plaintext. Continue?")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Println("Aborted.")
					return nil
				}
			}

			pageSize := boot.Config.Security.DBEncryption.CipherPageSize
			if pageSize <= 0 {
				pageSize = 4096
			}

			pass, err := prompt.Passphrase("Enter passphrase for DB decryption: ")
			if err != nil {
				return fmt.Errorf("read passphrase: %w", err)
			}

			fmt.Println("Decrypting database...")
			if err := dbmigrate.DecryptToPlaintext(dbPath, pass, pageSize); err != nil {
				return fmt.Errorf("db decryption: %w", err)
			}

			fmt.Println("Database decrypted successfully.")
			fmt.Println("Set security.dbEncryption.enabled=false in your config if you no longer want encryption.")
			return nil
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
