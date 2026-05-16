package passphrase

import (
	"errors"
	"fmt"
	"io"
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
	SourceKeyring                   // from hardware keyring (Touch ID / TPM)
)

// Options configures passphrase acquisition behavior.
type Options struct {
	KeyfilePath     string           // default: ~/.lango/keyfile
	AllowCreation   bool             // if true, prompt for confirmation on new passphrase
	KeyringProvider keyring.Provider // if non-nil, try secure keyring first (biometric/TPM)
}

// defaultKeyfilePath returns the default keyfile path (~/.lango/keyfile).
func defaultKeyfilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".lango", "keyfile"), nil
}

func acquireWithIO(opts Options, stdin io.Reader, stderr io.Writer, interactive bool) (string, Source, error) {
	keyfilePath := opts.KeyfilePath
	if keyfilePath == "" {
		var err error
		keyfilePath, err = defaultKeyfilePath()
		if err != nil {
			return "", 0, err
		}
	}

	if opts.KeyringProvider != nil {
		pass, err := opts.KeyringProvider.Get(keyring.Service, keyring.KeyMasterPassphrase)
		if err == nil && pass != "" {
			return pass, SourceKeyring, nil
		}
		if err != nil && !errors.Is(err, keyring.ErrNotFound) {
			fmt.Fprintf(stderr, "warning: keyring read failed: %v\n", err)
		}
	}

	if pass, err := ReadKeyfile(keyfilePath); err == nil {
		return pass, SourceKeyfile, nil
	}

	if interactive {
		pass, err := acquireInteractive(opts.AllowCreation)
		if err != nil {
			return "", 0, fmt.Errorf("interactive passphrase: %w", err)
		}
		return pass, SourceInteractive, nil
	}

	pass, err := ReadStdinPipeFromReader(stdin)
	if err != nil {
		return "", 0, fmt.Errorf("stdin passphrase: %w", err)
	}
	return pass, SourceStdin, nil
}

// Acquire obtains a passphrase from the highest-priority available source.
// Priority: keyring -> keyfile -> interactive terminal -> stdin pipe -> error
func Acquire(opts Options) (string, Source, error) {
	return acquireWithIO(opts, os.Stdin, os.Stderr, term.IsTerminal(int(syscall.Stdin)))
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
