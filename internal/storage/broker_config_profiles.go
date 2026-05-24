package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/storagebroker"
)

type brokerConfigProfiles struct {
	broker storagebroker.API
}

func NewBrokerConfigProfiles(broker storagebroker.API) ConfigProfileStore {
	if broker == nil {
		return nil
	}
	return &brokerConfigProfiles{broker: broker}
}

func (s *brokerConfigProfiles) Save(ctx context.Context, name string, cfg *config.Config, explicitKeys map[string]bool) error {
	return s.broker.ConfigSave(ctx, name, cfg, explicitKeys)
}

func (s *brokerConfigProfiles) Load(ctx context.Context, name string) (*config.Config, map[string]bool, error) {
	result, err := s.broker.ConfigLoad(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	var cfg config.Config
	if err := json.Unmarshal(result.Config, &cfg); err != nil {
		return nil, nil, fmt.Errorf("decode broker profile %q: %w", name, err)
	}
	return &cfg, result.ExplicitKeys, nil
}

func (s *brokerConfigProfiles) LoadActive(ctx context.Context) (string, *config.Config, map[string]bool, error) {
	result, err := s.broker.ConfigLoadActive(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	var cfg config.Config
	if err := json.Unmarshal(result.Config, &cfg); err != nil {
		return "", nil, nil, fmt.Errorf("decode active broker profile: %w", err)
	}
	return result.Name, &cfg, result.ExplicitKeys, nil
}

func (s *brokerConfigProfiles) SetActive(ctx context.Context, name string) error {
	return s.broker.ConfigSetActive(ctx, name)
}

func (s *brokerConfigProfiles) List(ctx context.Context) ([]configstore.ProfileInfo, error) {
	result, err := s.broker.ConfigList(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]configstore.ProfileInfo, 0, len(result.Profiles))
	for _, p := range result.Profiles {
		createdAt, _ := time.Parse(time.RFC3339, p.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, p.UpdatedAt)
		out = append(out, configstore.ProfileInfo{
			Name:      p.Name,
			Active:    p.Active,
			Version:   p.Version,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	return out, nil
}

func (s *brokerConfigProfiles) Delete(ctx context.Context, name string) error {
	return s.broker.ConfigDelete(ctx, name)
}

func (s *brokerConfigProfiles) Exists(ctx context.Context, name string) (bool, error) {
	result, err := s.broker.ConfigExists(ctx, name)
	if err != nil {
		return false, err
	}
	return result.Exists, nil
}
