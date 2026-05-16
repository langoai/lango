package passphrase

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/keyring"
)

func TestAcquireNonInteractive_Keyfile(t *testing.T) {
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "keyfile")
	want := "from-keyfile-pass"
	if err := WriteKeyfile(keyfilePath, want); err != nil {
		t.Fatalf("WriteKeyfile: %v", err)
	}

	got, src, err := AcquireNonInteractive(Options{KeyfilePath: keyfilePath})
	if err != nil {
		t.Fatalf("AcquireNonInteractive: %v", err)
	}
	if got != want {
		t.Fatalf("pass mismatch: got %q want %q", got, want)
	}
	if src != SourceKeyfile {
		t.Fatalf("expected SourceKeyfile, got %v", src)
	}
}

func TestAcquireNonInteractive_NoSource(t *testing.T) {
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent")

	_, _, err := AcquireNonInteractive(Options{KeyfilePath: keyfilePath})
	if err == nil {
		t.Fatal("expected error when neither keyring nor keyfile available")
	}
	if !errors.Is(err, ErrNoNonInteractiveSource) {
		t.Fatalf("expected ErrNoNonInteractiveSource, got %v", err)
	}
}

func TestAcquireNonInteractive_NeverPrompts(t *testing.T) {
	// Regression: AcquireNonInteractive must return quickly even without a tty
	// and without any available source. If the implementation slips in a
	// term.ReadPassword call, this test would block indefinitely. We use the
	// plain error check; if the function ever becomes interactive we'll notice
	// immediately in CI (tests will hang).
	dir := t.TempDir()
	_, _, err := AcquireNonInteractive(Options{KeyfilePath: filepath.Join(dir, "missing")})
	if !errors.Is(err, ErrNoNonInteractiveSource) {
		t.Fatalf("expected ErrNoNonInteractiveSource, got %v", err)
	}
}

type stubNonInteractiveKeyringProvider struct {
	pass string
	err  error
}

func (s stubNonInteractiveKeyringProvider) Get(service, key string) (string, error) {
	return s.pass, s.err
}
func (s stubNonInteractiveKeyringProvider) Set(service, key, value string) error { return nil }
func (s stubNonInteractiveKeyringProvider) Delete(service, key string) error     { return nil }

func TestAcquireNonInteractive_KeyringErrorWarnsAndFallsThrough(t *testing.T) {
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "keyfile")
	want := "from-keyfile-pass"
	if err := WriteKeyfile(keyfilePath, want); err != nil {
		t.Fatalf("WriteKeyfile: %v", err)
	}

	var errBuf bytes.Buffer
	got, src, err := acquireNonInteractiveWithIO(Options{
		KeyfilePath:     keyfilePath,
		KeyringProvider: stubNonInteractiveKeyringProvider{err: errors.New("boom")},
	}, &errBuf)
	if err != nil {
		t.Fatalf("acquireNonInteractiveWithIO: %v", err)
	}
	if got != want {
		t.Fatalf("pass mismatch: got %q want %q", got, want)
	}
	if src != SourceKeyfile {
		t.Fatalf("expected SourceKeyfile, got %v", src)
	}
	if errBuf.String() != "warning: keyring read failed: boom\n" {
		t.Fatalf("unexpected stderr warning: %q", errBuf.String())
	}
}

func TestAcquireNonInteractive_KeyringNotFoundDoesNotWarn(t *testing.T) {
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "keyfile")
	want := "from-keyfile-pass"
	if err := WriteKeyfile(keyfilePath, want); err != nil {
		t.Fatalf("WriteKeyfile: %v", err)
	}

	var errBuf bytes.Buffer
	got, src, err := acquireNonInteractiveWithIO(Options{
		KeyfilePath:     keyfilePath,
		KeyringProvider: stubNonInteractiveKeyringProvider{err: keyring.ErrNotFound},
	}, &errBuf)
	if err != nil {
		t.Fatalf("acquireNonInteractiveWithIO: %v", err)
	}
	if got != want {
		t.Fatalf("pass mismatch: got %q want %q", got, want)
	}
	if src != SourceKeyfile {
		t.Fatalf("expected SourceKeyfile, got %v", src)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected no stderr warning, got %q", errBuf.String())
	}
}
