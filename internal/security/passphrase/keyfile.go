package passphrase

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const requiredPerm fs.FileMode = 0600

// ReadKeyfile reads the passphrase from a keyfile, validating file permissions.
// Returns the passphrase with trailing whitespace trimmed.
func ReadKeyfile(path string) (string, error) {
	if err := ValidatePermissions(path); err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read keyfile: %w", err)
	}

	passphrase := strings.TrimRight(string(data), "\n\r \t")
	if passphrase == "" {
		return "", fmt.Errorf("keyfile is empty: %s", path)
	}

	return passphrase, nil
}

// WriteKeyfile creates a keyfile with 0600 permissions.
// Parent directories are created with 0700 permissions if needed.
func WriteKeyfile(path, passphrase string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create keyfile directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(passphrase+"\n"), requiredPerm); err != nil {
		return fmt.Errorf("write keyfile: %w", err)
	}

	return nil
}

// ShredKeyfile overwrites the keyfile content with zeros, syncs to disk, and removes it.
// Returns nil if the file does not exist (idempotent).
func ShredKeyfile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat keyfile for shred: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open keyfile for shred: %w", err)
	}

	zeros := make([]byte, info.Size())
	if _, err := f.Write(zeros); err != nil {
		f.Close()
		return fmt.Errorf("overwrite keyfile: %w", err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("sync keyfile: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close keyfile: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove keyfile: %w", err)
	}

	return nil
}

// ValidatePermissions checks that the file has exactly 0600 permissions.
func ValidatePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat keyfile: %w", err)
	}

	perm := info.Mode().Perm()
	if perm != requiredPerm {
		return fmt.Errorf(
			"keyfile %s has permissions %04o, want %04o",
			path, perm, requiredPerm,
		)
	}

	return nil
}
