package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// envelopeFileName is the fixed filename for the MasterKeyEnvelope inside the lango dir.
const envelopeFileName = "envelope.json"

// envelopeFilePerms is the required permission mask for the envelope file.
const envelopeFilePerms os.FileMode = 0o600

// EnvelopeFilePath returns the absolute path of the envelope file for the given lango dir.
func EnvelopeFilePath(langoDir string) string {
	return filepath.Join(langoDir, envelopeFileName)
}

// HasEnvelopeFile reports whether an envelope file exists under langoDir.
// I/O errors (other than not-exist) are reported as "exists=false" to keep the
// signature simple; callers that need to distinguish should use LoadEnvelopeFile.
func HasEnvelopeFile(langoDir string) bool {
	_, err := os.Stat(EnvelopeFilePath(langoDir))
	return err == nil
}

// LoadEnvelopeFile reads and parses <langoDir>/envelope.json.
// Returns (nil, nil) if the file does not exist (fresh install or legacy layout).
// Returns a wrapped ErrEnvelopeCorrupt on parse failure.
func LoadEnvelopeFile(langoDir string) (*MasterKeyEnvelope, error) {
	path := EnvelopeFilePath(langoDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read envelope: %w", err)
	}
	var env MasterKeyEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEnvelopeCorrupt, err)
	}
	if env.Version != EnvelopeVersion {
		return nil, fmt.Errorf("%w: unsupported version %d", ErrEnvelopeCorrupt, env.Version)
	}
	if len(env.Slots) == 0 {
		return nil, fmt.Errorf("%w: envelope has no KEK slots", ErrEnvelopeCorrupt)
	}
	return &env, nil
}

// StoreEnvelopeFile writes the envelope to <langoDir>/envelope.json atomically.
// The file is created with 0600 permissions. Uses write-to-temp-and-rename to avoid
// leaving a partial file on crash.
func StoreEnvelopeFile(langoDir string, env *MasterKeyEnvelope) error {
	if env == nil {
		return fmt.Errorf("store envelope: nil envelope")
	}
	if err := os.MkdirAll(langoDir, 0o700); err != nil {
		return fmt.Errorf("ensure lango dir: %w", err)
	}
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	path := EnvelopeFilePath(langoDir)
	tmp := path + ".tmp"
	// Remove stale temp file from a prior crash.
	_ = os.Remove(tmp)
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, envelopeFilePerms)
	if err != nil {
		return fmt.Errorf("open envelope temp: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write envelope temp: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("sync envelope temp: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close envelope temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename envelope: %w", err)
	}
	return nil
}
