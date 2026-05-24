package security

import (
	"fmt"
	"io"
)

func writeKeyringUpdateWarning(errWriter io.Writer, err error) {
	fmt.Fprintf(errWriter, "warning: keyring update failed: %v\n", err)
	fmt.Fprintf(errWriter, "  If a stale passphrase is stored, next headless bootstrap may fail.\n")
	fmt.Fprintf(errWriter, "  Fix: run `lango security keyring store` or clear the keyring entry manually.\n")
}
