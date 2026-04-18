package storage

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/storagebroker"
)

type brokerSecurityState struct {
	broker storagebroker.API
}

func NewBrokerSecurityState(broker storagebroker.API) SecurityStateStore {
	if broker == nil {
		return nil
	}
	return &brokerSecurityState{broker: broker}
}

func (s *brokerSecurityState) LoadSalt() ([]byte, error) {
	state, err := s.broker.LoadSecurityState(context.Background())
	if err != nil {
		return nil, err
	}
	return state.Salt, nil
}

func (s *brokerSecurityState) StoreSalt(salt []byte) error {
	return s.broker.StoreSalt(context.Background(), salt)
}

func (s *brokerSecurityState) LoadChecksum() ([]byte, error) {
	state, err := s.broker.LoadSecurityState(context.Background())
	if err != nil {
		return nil, err
	}
	return state.Checksum, nil
}

func (s *brokerSecurityState) StoreChecksum(checksum []byte) error {
	return s.broker.StoreChecksum(context.Background(), checksum)
}

func (s *brokerSecurityState) IsFirstRun() (bool, error) {
	state, err := s.broker.LoadSecurityState(context.Background())
	if err != nil {
		return false, fmt.Errorf("load broker security state: %w", err)
	}
	return state.FirstRun, nil
}
