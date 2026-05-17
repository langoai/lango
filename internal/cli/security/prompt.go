package security

import (
	"io"

	"github.com/langoai/lango/internal/cli/prompt"
)

var (
	securityRequireInteractiveInput = prompt.RequireInteractiveInput
	securityPassphrase              = prompt.PassphraseIO
	securityPassphraseConfirm       = func(out io.Writer, promptText, confirmPrompt string) (string, error) {
		return prompt.PassphraseConfirmIO(out, promptText, confirmPrompt)
	}
)
