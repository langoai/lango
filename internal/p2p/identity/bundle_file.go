package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const bundleFileName = "identity-bundle.json"
const knownBundlesDir = "known-bundles"
const bundleFilePerms os.FileMode = 0o600

// BundleFilePath returns the path to the local identity bundle file.
func BundleFilePath(langoDir string) string {
	return filepath.Join(langoDir, bundleFileName)
}

// HasBundleFile reports whether a local identity bundle file exists.
func HasBundleFile(langoDir string) bool {
	_, err := os.Stat(BundleFilePath(langoDir))
	return err == nil
}

// LoadBundleFile reads and parses the local identity bundle.
// Returns (nil, nil) if the file does not exist.
func LoadBundleFile(langoDir string) (*IdentityBundle, error) {
	path := BundleFilePath(langoDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read identity bundle: %w", err)
	}
	var bundle IdentityBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("parse identity bundle: %w", err)
	}
	return &bundle, nil
}

// StoreBundleFile writes the identity bundle atomically.
// Uses write-to-temp-and-rename to avoid partial files on crash.
func StoreBundleFile(langoDir string, bundle *IdentityBundle) error {
	if bundle == nil {
		return fmt.Errorf("store identity bundle: nil bundle")
	}
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal identity bundle: %w", err)
	}

	path := BundleFilePath(langoDir)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, bundleFilePerms); err != nil {
		return fmt.Errorf("write identity bundle temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename identity bundle: %w", err)
	}
	return nil
}

// knownBundlePath returns the path for a remote peer's cached bundle.
func knownBundlePath(langoDir, didV2 string) string {
	// Use a hash of the DID to avoid filesystem issues with special chars.
	safe := filepath.Base(didV2) // safe filename from DID
	return filepath.Join(langoDir, knownBundlesDir, safe+".json")
}

// StoreKnownBundle persists a remote peer's IdentityBundle to disk.
func StoreKnownBundle(langoDir string, didV2 string, bundle *IdentityBundle) error {
	if bundle == nil {
		return fmt.Errorf("store known bundle: nil bundle")
	}
	dir := filepath.Join(langoDir, knownBundlesDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create known-bundles dir: %w", err)
	}
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal known bundle: %w", err)
	}
	path := knownBundlePath(langoDir, didV2)
	return os.WriteFile(path, data, bundleFilePerms)
}

// LoadKnownBundle loads a cached remote peer bundle from disk.
// Returns (nil, nil) if not found.
func LoadKnownBundle(langoDir string, didV2 string) (*IdentityBundle, error) {
	path := knownBundlePath(langoDir, didV2)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read known bundle: %w", err)
	}
	var bundle IdentityBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("parse known bundle: %w", err)
	}
	return &bundle, nil
}
