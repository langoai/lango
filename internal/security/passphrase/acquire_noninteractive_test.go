package passphrase

import (
	"errors"
	"path/filepath"
	"testing"
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
