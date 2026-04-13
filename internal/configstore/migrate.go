package configstore

import (
	"context"
	"fmt"
	"github.com/langoai/lango/internal/logging"
	"os"

	"github.com/langoai/lango/internal/config"
)

// MigrateFromJSON reads a JSON config file and imports it as an encrypted profile.
// The imported profile is set as the active profile.
func MigrateFromJSON(ctx context.Context, store *Store, jsonPath, profileName string) error {
	result, err := config.Load(jsonPath)
	if err != nil {
		return fmt.Errorf("load config from %q: %w", jsonPath, err)
	}

	if err := store.Save(ctx, profileName, result.Config, result.ExplicitKeys); err != nil {
		return fmt.Errorf("save profile %q: %w", profileName, err)
	}

	if err := store.SetActive(ctx, profileName); err != nil {
		return fmt.Errorf("set active profile %q: %w", profileName, err)
	}

	// Delete the source JSON file after successful import for security.
	if err := os.Remove(jsonPath); err != nil {
		logging.SubsystemSugar("configstore").Warnw("imported but could not delete source", "path", jsonPath, "error", err)
	}

	return nil
}
