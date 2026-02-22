package configstore

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/langoai/lango/internal/config"
)

// MigrateFromJSON reads a JSON config file and imports it as an encrypted profile.
// The imported profile is set as the active profile.
func MigrateFromJSON(ctx context.Context, store *Store, jsonPath, profileName string) error {
	cfg, err := config.Load(jsonPath)
	if err != nil {
		return fmt.Errorf("load config from %q: %w", jsonPath, err)
	}

	if err := store.Save(ctx, profileName, cfg); err != nil {
		return fmt.Errorf("save profile %q: %w", profileName, err)
	}

	if err := store.SetActive(ctx, profileName); err != nil {
		return fmt.Errorf("set active profile %q: %w", profileName, err)
	}

	// Delete the source JSON file after successful import for security.
	if err := os.Remove(jsonPath); err != nil {
		log.Printf("WARNING: imported successfully but could not delete source file %q: %v", jsonPath, err)
	}

	return nil
}
