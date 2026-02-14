package configstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/configprofile"
	"github.com/langowarny/lango/internal/security"
)

// Sentinel errors for configstore operations.
var (
	ErrProfileNotFound = errors.New("profile not found")
	ErrNoActiveProfile = errors.New("no active profile")
	ErrDeleteActive    = errors.New("cannot delete the active profile")
)

// Store manages encrypted configuration profiles.
type Store struct {
	client *ent.Client
	crypto security.CryptoProvider
}

// NewStore creates a new configuration store.
func NewStore(client *ent.Client, crypto security.CryptoProvider) *Store {
	return &Store{
		client: client,
		crypto: crypto,
	}
}

// Save serializes the config to JSON, encrypts it, and stores it as a profile.
// If a profile with the given name already exists, it is updated.
func (s *Store) Save(ctx context.Context, name string, cfg *config.Config) error {
	plainJSON, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	encrypted, err := s.crypto.Encrypt(ctx, "local", plainJSON)
	if err != nil {
		return fmt.Errorf("encrypt config: %w", err)
	}

	// Check if profile already exists.
	existing, err := s.client.ConfigProfile.
		Query().
		Where(configprofile.NameEQ(name)).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("query profile %q: %w", name, err)
	}

	if existing != nil {
		_, err = existing.Update().
			SetEncryptedData(encrypted).
			AddVersion(1).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("update profile %q: %w", name, err)
		}
		return nil
	}

	_, err = s.client.ConfigProfile.
		Create().
		SetName(name).
		SetEncryptedData(encrypted).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create profile %q: %w", name, err)
	}

	return nil
}

// Load decrypts and deserializes a named profile's configuration.
func (s *Store) Load(ctx context.Context, name string) (*config.Config, error) {
	profile, err := s.client.ConfigProfile.
		Query().
		Where(configprofile.NameEQ(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("query profile %q: %w", name, err)
	}

	plainJSON, err := s.crypto.Decrypt(ctx, "local", profile.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("decrypt profile %q: %w", name, err)
	}

	var cfg config.Config
	if err := json.Unmarshal(plainJSON, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal profile %q: %w", name, err)
	}

	return &cfg, nil
}

// LoadActive loads the currently active profile's configuration.
// Returns the profile name, config, and any error.
func (s *Store) LoadActive(ctx context.Context) (string, *config.Config, error) {
	profile, err := s.client.ConfigProfile.
		Query().
		Where(configprofile.ActiveEQ(true)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil, ErrNoActiveProfile
		}
		return "", nil, fmt.Errorf("query active profile: %w", err)
	}

	plainJSON, err := s.crypto.Decrypt(ctx, "local", profile.EncryptedData)
	if err != nil {
		return "", nil, fmt.Errorf("decrypt active profile: %w", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(plainJSON, &cfg); err != nil {
		return "", nil, fmt.Errorf("unmarshal active profile: %w", err)
	}

	return profile.Name, &cfg, nil
}

// SetActive deactivates all profiles and activates the named profile.
func (s *Store) SetActive(ctx context.Context, name string) error {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Deactivate all profiles.
	_, err = tx.ConfigProfile.
		Update().
		Where(configprofile.ActiveEQ(true)).
		SetActive(false).
		Save(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("deactivate profiles: %w", err))
	}

	// Activate the named profile.
	n, err := tx.ConfigProfile.
		Update().
		Where(configprofile.NameEQ(name)).
		SetActive(true).
		Save(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("activate profile %q: %w", name, err))
	}
	if n == 0 {
		return rollback(tx, ErrProfileNotFound)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// List returns metadata for all profiles (no decryption needed).
func (s *Store) List(ctx context.Context) ([]ProfileInfo, error) {
	profiles, err := s.client.ConfigProfile.
		Query().
		Order(configprofile.ByName()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}

	infos := make([]ProfileInfo, len(profiles))
	for i, p := range profiles {
		infos[i] = ProfileInfo{
			Name:      p.Name,
			Active:    p.Active,
			Version:   p.Version,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}

	return infos, nil
}

// Delete removes a profile by name.
// Returns an error if the profile is currently active.
func (s *Store) Delete(ctx context.Context, name string) error {
	profile, err := s.client.ConfigProfile.
		Query().
		Where(configprofile.NameEQ(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrProfileNotFound
		}
		return fmt.Errorf("query profile %q: %w", name, err)
	}

	if profile.Active {
		return ErrDeleteActive
	}

	if err := s.client.ConfigProfile.DeleteOne(profile).Exec(ctx); err != nil {
		return fmt.Errorf("delete profile %q: %w", name, err)
	}

	return nil
}

// Exists checks if a profile with the given name exists.
func (s *Store) Exists(ctx context.Context, name string) (bool, error) {
	exists, err := s.client.ConfigProfile.
		Query().
		Where(configprofile.NameEQ(name)).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("check profile %q: %w", name, err)
	}

	return exists, nil
}

// rollback calls tx.Rollback and wraps the original error.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		return fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
	}
	return err
}
