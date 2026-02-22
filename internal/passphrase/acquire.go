package passphrase

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/langoai/lango/internal/cli/prompt"
	"golang.org/x/term"
)

// Source represents how the passphrase was obtained.
type Source int

const (
	SourceKeyfile     Source = iota // from ~/.lango/keyfile
	SourceInteractive               // from interactive terminal prompt
	SourceStdin                     // from piped stdin
)

// Options configures passphrase acquisition behavior.
type Options struct {
	KeyfilePath   string // default: ~/.lango/keyfile
	AllowCreation bool   // if true, prompt for confirmation on new passphrase
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
// Priority: keyfile -> interactive terminal -> stdin pipe -> error
func Acquire(opts Options) (string, Source, error) {
	keyfilePath := opts.KeyfilePath
	if keyfilePath == "" {
		var err error
		keyfilePath, err = defaultKeyfilePath()
		if err != nil {
			return "", 0, err
		}
	}

	// 1. Try keyfile
	if pass, err := ReadKeyfile(keyfilePath); err == nil {
		return pass, SourceKeyfile, nil
	}

	// 2. Try interactive terminal
	if term.IsTerminal(int(syscall.Stdin)) {
		pass, err := acquireInteractive(opts.AllowCreation)
		if err != nil {
			return "", 0, fmt.Errorf("interactive passphrase: %w", err)
		}
		return pass, SourceInteractive, nil
	}

	// 3. Try stdin pipe
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
