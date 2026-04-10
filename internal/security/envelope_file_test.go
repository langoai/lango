package security

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnvelopeFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)

	if err := StoreEnvelopeFile(dir, env); err != nil {
		t.Fatalf("StoreEnvelopeFile: %v", err)
	}
	if !HasEnvelopeFile(dir) {
		t.Fatal("HasEnvelopeFile should report true after store")
	}

	got, err := LoadEnvelopeFile(dir)
	if err != nil {
		t.Fatalf("LoadEnvelopeFile: %v", err)
	}
	if got == nil {
		t.Fatal("loaded envelope is nil")
	}
	if got.Version != env.Version {
		t.Fatalf("version mismatch: got %d want %d", got.Version, env.Version)
	}
	if len(got.Slots) != len(env.Slots) {
		t.Fatalf("slot count mismatch: got %d want %d", len(got.Slots), len(env.Slots))
	}
	// Round-trip the passphrase unwrap to verify slot data survived JSON.
	unwrapped, _, err := got.UnwrapFromPassphrase(testPassphrase)
	if err != nil {
		t.Fatalf("unwrap after reload: %v", err)
	}
	ZeroBytes(unwrapped)
}

func TestEnvelopeFile_MissingReturnsNil(t *testing.T) {
	dir := t.TempDir()
	got, err := LoadEnvelopeFile(dir)
	if err != nil {
		t.Fatalf("LoadEnvelopeFile on missing file: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil envelope for missing file")
	}
	if HasEnvelopeFile(dir) {
		t.Fatal("HasEnvelopeFile should report false on empty dir")
	}
}

func TestEnvelopeFile_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions not applicable on windows")
	}
	dir := t.TempDir()
	env, mk, _ := NewEnvelope(testPassphrase)
	defer ZeroBytes(mk)
	if err := StoreEnvelopeFile(dir, env); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(EnvelopeFilePath(dir))
	if err != nil {
		t.Fatal(err)
	}
	mode := info.Mode().Perm()
	if mode != envelopeFilePerms {
		t.Fatalf("expected perms %o, got %o", envelopeFilePerms, mode)
	}
}

func TestEnvelopeFile_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := EnvelopeFilePath(dir)
	if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadEnvelopeFile(dir)
	if !errors.Is(err, ErrEnvelopeCorrupt) {
		t.Fatalf("expected ErrEnvelopeCorrupt, got %v", err)
	}
}

func TestEnvelopeFile_UnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	path := EnvelopeFilePath(dir)
	if err := os.WriteFile(path, []byte(`{"version":99,"slots":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadEnvelopeFile(dir)
	if !errors.Is(err, ErrEnvelopeCorrupt) {
		t.Fatalf("expected ErrEnvelopeCorrupt, got %v", err)
	}
}

func TestEnvelopeFile_AtomicRename(t *testing.T) {
	// Ensure the temp file is cleaned up after a successful store.
	dir := t.TempDir()
	env, mk, _ := NewEnvelope(testPassphrase)
	defer ZeroBytes(mk)
	if err := StoreEnvelopeFile(dir, env); err != nil {
		t.Fatal(err)
	}
	tmp := EnvelopeFilePath(dir) + ".tmp"
	if _, err := os.Stat(tmp); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected temp file %q to be gone, stat err=%v", filepath.Base(tmp), err)
	}
}
