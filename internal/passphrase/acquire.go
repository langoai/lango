package passphrase

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/keyring"
	"golang.org/x/term"
)

// Source represents how the passphrase was obtained.
type Source int

const (
	SourceKeyfile     Source = iota // from ~/.lango/keyfile
	SourceInteractive               // from interactive terminal prompt
	SourceStdin                     // from piped stdin
	SourceKeyring                   // from OS keyring (macOS Keychain / Linux secret-service / Windows DPAPI)
)

// Options configures passphrase acquisition behavior.
type Options struct {
	KeyfilePath     string           // default: ~/.lango/keyfile
	AllowCreation   bool             // if true, prompt for confirmation on new passphrase
	KeyringProvider keyring.Provider // if non-nil, try OS keyring first
}

// defaultKeyfilePath returns the default keyfile path (~/.lango/keyfile).
func defaultKeyfilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".lango", "keyfile"), nil
}

// Acquire obtains a passphrase from the highest-priority available source.
// Priority: keyring -> keyfile -> interactive terminal -> stdin pipe -> error
func Acquire(opts Options) (string, Source, error) {
	keyfilePath := opts.KeyfilePath
	if keyfilePath == "" {
		var err error
		keyfilePath, err = defaultKeyfilePath()
		if err != nil {
			return "", 0, err
		}
	}

	// 1. Try OS keyring (highest priority).
	if opts.KeyringProvider != nil {
		pass, err := opts.KeyringProvider.Get(keyring.Service, keyring.KeyMasterPassphrase)
		if err == nil && pass != "" {
			return pass, SourceKeyring, nil
		}
		// Silently fall through on ErrNotFound or any other keyring error.
		if err != nil && !errors.Is(err, keyring.ErrNotFound) {
			// Log-worthy but not fatal — continue to next source.
		}
	}

	// 2. Try keyfile.
	if pass, err := ReadKeyfile(keyfilePath); err == nil {
		return pass, SourceKeyfile, nil
	}

	// 3. Try interactive terminal.
	if term.IsTerminal(int(syscall.Stdin)) {
		pass, err := acquireInteractive(opts.AllowCreation)
		if err != nil {
			return "", 0, fmt.Errorf("interactive passphrase: %w", err)
		}
		return pass, SourceInteractive, nil
	}

	// 4. Try stdin pipe.
	pass, err := ReadStdinPipe()
	if err != nil {
		return "", 0, fmt.Errorf("stdin passphrase: %w", err)
	}
	return pass, SourceStdin, nil
}

// acquireInteractive prompts the user for a passphrase via the terminal.
func acquireInteractive(allowCreation bool) (string, error) {
	if allowCreation {
		return prompt.PassphraseConfirm(
			"Enter new passphrase: ",
			"Confirm passphrase: ",
		)
	}
	return prompt.Passphrase("Enter passphrase: ")
}
