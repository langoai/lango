package prompt

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/langoai/lango/internal/lineio"
)

var (
	passphraseOutput       io.Writer = os.Stdout
	passphraseInputFD                = func() int { return int(syscall.Stdin) }
	passphraseReadPassword           = term.ReadPassword
	confirmInput           io.Reader = os.Stdin
	confirmOutput          io.Writer = os.Stdout
	interactiveCheck                 = func() bool { return term.IsTerminal(int(syscall.Stdin)) }
)

// IsInteractive returns true if the standard input is a terminal
func IsInteractive() bool {
	return interactiveCheck()
}

// RequireInteractiveTerminal fails when the current stdin is not interactive.
func RequireInteractiveTerminal(message string) error {
	if !IsInteractive() {
		return errors.New(message)
	}
	return nil
}

// Passphrase prompts the user for a passphrase with hidden input
func Passphrase(prompt string) (string, error) {
	fmt.Fprint(passphraseOutput, prompt)
	bytePassword, err := passphraseReadPassword(passphraseInputFD())
	fmt.Fprintln(passphraseOutput) // Add newline after input
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

// Confirm prompts the user for a yes/no confirmation and returns true for yes.
func Confirm(msg string) (bool, error) {
	return ConfirmDenyOnEOFIO(confirmInput, confirmOutput, msg)
}

// ConfirmIO prompts the user for a yes/no confirmation using the provided streams.
func ConfirmIO(in io.Reader, out io.Writer, msg string) (bool, error) {
	line, err := ReadLineIO(in, out, fmt.Sprintf("%s [y/N]: ", msg))
	if err != nil {
		return false, err
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes", nil
}

// ConfirmTTYIO rejects non-terminal file input with the provided guidance error
// and otherwise delegates to ConfirmIO. EOF is treated as a clean denial.
func ConfirmTTYIO(in io.Reader, out io.Writer, msg, nonTTYError string) (bool, error) {
	if err := RequireTTYInput(in, nonTTYError); err != nil {
		return false, err
	}
	return ConfirmDenyOnEOFIO(in, out, msg)
}

// RequireTTYInput rejects non-terminal file input with the provided guidance
// error. Non-file readers are allowed so tests can inject explicit streams.
func RequireTTYInput(in io.Reader, nonTTYError string) error {
	if f, ok := in.(*os.File); ok && !term.IsTerminal(int(f.Fd())) {
		return errors.New(nonTTYError)
	}
	return nil
}

// ConfirmDenyOnEOFIO maps EOF on the confirmation input to a clean denial.
func ConfirmDenyOnEOFIO(in io.Reader, out io.Writer, msg string) (bool, error) {
	ok, err := ConfirmIO(in, out, msg)
	if errors.Is(err, io.EOF) {
		return false, nil
	}
	return ok, err
}

// ReadLineIO writes a visible prompt and returns one line from the supplied input.
func ReadLineIO(in io.Reader, out io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return "", err
	}
	return lineio.ReadLine(in)
}

// PassphraseConfirm prompts for a passphrase and its confirmation
func PassphraseConfirm(prompt, confirmPrompt string) (string, error) {
	pass1, err := Passphrase(prompt)
	if err != nil {
		return "", err
	}

	pass2, err := Passphrase(confirmPrompt)
	if err != nil {
		return "", err
	}

	if pass1 != pass2 {
		return "", fmt.Errorf("passphrases do not match")
	}

	return pass1, nil
}
