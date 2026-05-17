package prompt

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestPassphrase_UsesSeams(t *testing.T) {
	origOut := passphraseOutput
	origFD := passphraseInputFD
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseOutput = origOut
		passphraseInputFD = origFD
		passphraseReadPassword = origRead
	})

	var out bytes.Buffer
	passphraseOutput = &out
	passphraseInputFD = func() int { return 42 }
	passphraseReadPassword = func(fd int) ([]byte, error) {
		if fd != 42 {
			t.Fatalf("unexpected fd: %d", fd)
		}
		return []byte("secret-pass"), nil
	}

	got, err := Passphrase("Enter passphrase: ")
	if err != nil {
		t.Fatalf("Passphrase returned error: %v", err)
	}
	if got != "secret-pass" {
		t.Fatalf("expected secret-pass, got %q", got)
	}
	if out.String() != "Enter passphrase: \n" {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestPassphrase_PropagatesReadError(t *testing.T) {
	origOut := passphraseOutput
	origFD := passphraseInputFD
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseOutput = origOut
		passphraseInputFD = origFD
		passphraseReadPassword = origRead
	})

	passphraseOutput = &bytes.Buffer{}
	passphraseInputFD = func() int { return 0 }
	passphraseReadPassword = func(fd int) ([]byte, error) {
		return nil, errors.New("read failed")
	}

	_, err := Passphrase("Enter passphrase: ")
	if err == nil {
		t.Fatal("expected read error")
	}
}

func TestPassphraseIO_RoutesPromptToExplicitWriter(t *testing.T) {
	origFD := passphraseInputFD
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseInputFD = origFD
		passphraseReadPassword = origRead
	})

	var out bytes.Buffer
	passphraseInputFD = func() int { return 42 }
	passphraseReadPassword = func(fd int) ([]byte, error) {
		if fd != 42 {
			t.Fatalf("unexpected fd: %d", fd)
		}
		return []byte("secret-value"), nil
	}

	got, err := PassphraseIO(&out, "Enter secret value: ")

	if err != nil {
		t.Fatalf("PassphraseIO returned error: %v", err)
	}
	if got != "secret-value" {
		t.Fatalf("expected secret-value, got %q", got)
	}
	if out.String() != "Enter secret value: \n" {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestPassphraseIO_ReturnsPromptWriteError(t *testing.T) {
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseReadPassword = origRead
	})

	passphraseReadPassword = func(fd int) ([]byte, error) {
		t.Fatal("password reader should not run when prompt write fails")
		return nil, nil
	}

	_, err := PassphraseIO(failingWriter{}, "Enter secret value: ")

	if err == nil {
		t.Fatal("expected prompt write error")
	}
}

func TestPassphraseConfirmIO_RoutesBothPromptsToExplicitWriter(t *testing.T) {
	origFD := passphraseInputFD
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseInputFD = origFD
		passphraseReadPassword = origRead
	})

	var out bytes.Buffer
	var calls int
	passphraseInputFD = func() int { return 42 }
	passphraseReadPassword = func(fd int) ([]byte, error) {
		if fd != 42 {
			t.Fatalf("unexpected fd: %d", fd)
		}
		calls++
		return []byte("same-passphrase"), nil
	}

	got, err := PassphraseConfirmIO(&out, "Enter NEW passphrase: ", "Confirm NEW passphrase: ")

	if err != nil {
		t.Fatalf("PassphraseConfirmIO returned error: %v", err)
	}
	if got != "same-passphrase" {
		t.Fatalf("expected same-passphrase, got %q", got)
	}
	if calls != 2 {
		t.Fatalf("expected 2 password reads, got %d", calls)
	}
	if out.String() != "Enter NEW passphrase: \nConfirm NEW passphrase: \n" {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestPassphraseConfirmIO_Mismatch(t *testing.T) {
	origFD := passphraseInputFD
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseInputFD = origFD
		passphraseReadPassword = origRead
	})

	var calls int
	var out bytes.Buffer
	passphraseInputFD = func() int { return 42 }
	passphraseReadPassword = func(fd int) ([]byte, error) {
		calls++
		if calls == 1 {
			return []byte("first"), nil
		}
		return []byte("second"), nil
	}

	_, err := PassphraseConfirmIO(&out, "Enter: ", "Confirm: ")

	if err == nil {
		t.Fatal("expected mismatch error")
	}
	if out.String() != "Enter: \nConfirm: \n" {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestPassphraseConfirmIO_ReturnsPromptWriteError(t *testing.T) {
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseReadPassword = origRead
	})

	passphraseReadPassword = func(fd int) ([]byte, error) {
		t.Fatal("password reader should not run when prompt write fails")
		return nil, nil
	}

	_, err := PassphraseConfirmIO(failingWriter{}, "Enter: ", "Confirm: ")

	if err == nil {
		t.Fatal("expected prompt write error")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestPassphraseConfirm_Mismatch(t *testing.T) {
	origOut := passphraseOutput
	origFD := passphraseInputFD
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseOutput = origOut
		passphraseInputFD = origFD
		passphraseReadPassword = origRead
	})

	var calls int
	passphraseOutput = &bytes.Buffer{}
	passphraseInputFD = func() int { return 0 }
	passphraseReadPassword = func(fd int) ([]byte, error) {
		calls++
		if calls == 1 {
			return []byte("first"), nil
		}
		return []byte("second"), nil
	}

	_, err := PassphraseConfirm("Enter: ", "Confirm: ")
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestPassphraseConfirm_Success(t *testing.T) {
	origOut := passphraseOutput
	origFD := passphraseInputFD
	origRead := passphraseReadPassword
	t.Cleanup(func() {
		passphraseOutput = origOut
		passphraseInputFD = origFD
		passphraseReadPassword = origRead
	})

	var calls int
	var out bytes.Buffer
	passphraseOutput = &out
	passphraseInputFD = func() int { return 0 }
	passphraseReadPassword = func(fd int) ([]byte, error) {
		calls++
		return []byte("same-passphrase"), nil
	}

	got, err := PassphraseConfirm("Enter: ", "Confirm: ")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if got != "same-passphrase" {
		t.Fatalf("expected same-passphrase, got %q", got)
	}
	if calls != 2 {
		t.Fatalf("expected 2 password reads, got %d", calls)
	}
	if out.String() != "Enter: \nConfirm: \n" {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestConfirmIO_Yes(t *testing.T) {
	var out bytes.Buffer
	ok, err := ConfirmIO(bytes.NewBufferString("yes\n"), &out, "Allow action?")
	if err != nil {
		t.Fatalf("ConfirmIO returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected approval")
	}
	if out.String() != "Allow action? [y/N]: " {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestConfirmIO_No(t *testing.T) {
	var out bytes.Buffer
	ok, err := ConfirmIO(bytes.NewBufferString("n\n"), &out, "Allow action?")
	if err != nil {
		t.Fatalf("ConfirmIO returned error: %v", err)
	}
	if ok {
		t.Fatal("expected denial")
	}
}

func TestConfirmIO_ReadError(t *testing.T) {
	var out bytes.Buffer
	_, err := ConfirmIO(bytes.NewBuffer(nil), &out, "Allow action?")
	if err == nil {
		t.Fatal("expected read error")
	}
}

func TestReadLineIO_UsesInjectedStreams(t *testing.T) {
	var out bytes.Buffer
	line, err := ReadLineIO(bytes.NewBufferString("alpha\n"), &out, "Enter word 1 to confirm: ")

	if err != nil {
		t.Fatalf("ReadLineIO returned error: %v", err)
	}
	if line != "alpha\n" {
		t.Fatalf("expected raw line with newline, got %q", line)
	}
	if out.String() != "Enter word 1 to confirm: " {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestReadLineIO_ReadError(t *testing.T) {
	var out bytes.Buffer
	_, err := ReadLineIO(bytes.NewBuffer(nil), &out, "Enter value: ")
	if err == nil {
		t.Fatal("expected read error")
	}
}

func TestConfirmTTYIO_RejectsNonTTYInput(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	t.Cleanup(func() { _ = reader.Close() })
	t.Cleanup(func() { _ = writer.Close() })

	var out bytes.Buffer
	ok, err := ConfirmTTYIO(reader, &out, "Install this pack?", "stdin is not a TTY; pass --yes for scripted runs")
	if err == nil {
		t.Fatal("expected non-tty guidance error")
	}
	if ok {
		t.Fatal("expected denial for non-tty input")
	}
	if out.String() != "" {
		t.Fatalf("expected no prompt output, got %q", out.String())
	}
}

func TestRequireTTYInput_RejectsNonTTYFile(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	t.Cleanup(func() { _ = reader.Close() })
	t.Cleanup(func() { _ = writer.Close() })

	err = RequireTTYInput(reader, "use --force for non-interactive mode")
	if err == nil {
		t.Fatal("expected non-tty guidance error")
	}
	if err.Error() != "use --force for non-interactive mode" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequireTTYInput_AllowsNonFileReader(t *testing.T) {
	err := RequireTTYInput(bytes.NewBufferString("y\n"), "use --force for non-interactive mode")
	if err != nil {
		t.Fatalf("expected non-file reader to pass, got error: %v", err)
	}
}

func TestRequireInteractiveTerminal_UsesInjectedInteractiveState(t *testing.T) {
	origInteractive := interactiveCheck
	t.Cleanup(func() {
		interactiveCheck = origInteractive
	})

	interactiveCheck = func() bool { return false }
	err := RequireInteractiveTerminal("this command requires an interactive terminal")
	if err == nil {
		t.Fatal("expected interactive terminal error")
	}
	if err.Error() != "this command requires an interactive terminal" {
		t.Fatalf("unexpected error: %v", err)
	}

	interactiveCheck = func() bool { return true }
	err = RequireInteractiveTerminal("this command requires an interactive terminal")
	if err != nil {
		t.Fatalf("expected nil when interactive, got error: %v", err)
	}
}

func TestConfirmDenyOnEOFIO_TreatsEOFAsDenial(t *testing.T) {
	var out bytes.Buffer
	ok, err := ConfirmDenyOnEOFIO(bytes.NewBuffer(nil), &out, "Continue?")
	if err != nil {
		t.Fatalf("expected EOF denial, got error: %v", err)
	}
	if ok {
		t.Fatal("expected denial on EOF")
	}
	if out.String() != "Continue? [y/N]: " {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestConfirmTTYIO_TreatsEOFAsDenial(t *testing.T) {
	var out bytes.Buffer
	ok, err := ConfirmTTYIO(bytes.NewBuffer(nil), &out, "Retry operation?", "")
	if err != nil {
		t.Fatalf("expected EOF denial, got error: %v", err)
	}
	if ok {
		t.Fatal("expected denial on EOF")
	}
	if out.String() != "Retry operation? [y/N]: " {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestConfirm_UsesInjectedStreamsForApproval(t *testing.T) {
	origIn := confirmInput
	origOut := confirmOutput
	t.Cleanup(func() {
		confirmInput = origIn
		confirmOutput = origOut
	})

	var out bytes.Buffer
	confirmInput = bytes.NewBufferString("y\n")
	confirmOutput = &out

	ok, err := Confirm("Continue?")
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected approval")
	}
	if out.String() != "Continue? [y/N]: " {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}

func TestConfirm_UsesInjectedStreamsForDenial(t *testing.T) {
	origIn := confirmInput
	origOut := confirmOutput
	t.Cleanup(func() {
		confirmInput = origIn
		confirmOutput = origOut
	})

	confirmInput = bytes.NewBufferString("no\n")
	confirmOutput = &bytes.Buffer{}

	ok, err := Confirm("Continue?")
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if ok {
		t.Fatal("expected denial")
	}
}

func TestConfirm_UsesInjectedStreamsForEOFDenial(t *testing.T) {
	origIn := confirmInput
	origOut := confirmOutput
	t.Cleanup(func() {
		confirmInput = origIn
		confirmOutput = origOut
	})

	confirmInput = bytes.NewBuffer(nil)
	confirmOutput = &bytes.Buffer{}

	ok, err := Confirm("Continue?")
	if err != nil {
		t.Fatalf("expected EOF denial, got error: %v", err)
	}
	if ok {
		t.Fatal("expected denial on EOF")
	}
	if got := confirmOutput.(*bytes.Buffer).String(); got != "Continue? [y/N]: " {
		t.Fatalf("unexpected prompt output: %q", got)
	}
}
