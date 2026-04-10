package passphrase

import (
	"errors"
	"fmt"
	"os"

	"github.com/langoai/lango/internal/keyring"
)

// ErrNoNonInteractiveSource is returned by AcquireNonInteractive when neither
// the keyring nor a keyfile can provide a passphrase. Callers that see this
// error should gracefully degrade (show zero counts, skip DB sections, etc.)
// instead of propagating the error to the user.
var ErrNoNonInteractiveSource = errors.New("no non-interactive passphrase source available")

// AcquireNonInteractive obtains a passphrase WITHOUT triggering any interactive
// terminal prompt or stdin pipe read. It is used by commands that must work
// in non-interactive environments (e.g. `lango security status` default path)
// while still being able to read stored state when a passphrase is available.
//
// Priority:
//  1. Secure keyring (Touch ID / TPM) — OS-level biometric prompts are
//     permitted because they are not CLI interactive prompts. TPM reads are
//     silent.
//  2. Keyfile at opts.KeyfilePath (or the default ~/.lango/keyfile).
//
// If neither source yields a passphrase, the function returns
// ErrNoNonInteractiveSource. It SHALL NEVER call term.ReadPassword or read
// from os.Stdin pipe. It SHALL NEVER block on user input beyond whatever the
// keyring provider itself does (e.g., a Touch ID prompt).
func AcquireNonInteractive(opts Options) (string, Source, error) {
	keyfilePath := opts.KeyfilePath
	if keyfilePath == "" {
		var err error
		keyfilePath, err = defaultKeyfilePath()
		if err != nil {
			return "", 0, err
		}
	}

	// 1. Try secure keyring first. A nil provider skips this step silently.
	if opts.KeyringProvider != nil {
		pass, err := opts.KeyringProvider.Get(keyring.Service, keyring.KeyMasterPassphrase)
		if err == nil && pass != "" {
			return pass, SourceKeyring, nil
		}
		// Soft-fail: not-found is normal; other errors fall through with a warning.
		if err != nil && !errors.Is(err, keyring.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "warning: keyring read failed: %v\n", err)
		}
	}

	// 2. Try keyfile.
	if pass, err := ReadKeyfile(keyfilePath); err == nil {
		return pass, SourceKeyfile, nil
	}

	return "", 0, ErrNoNonInteractiveSource
}
