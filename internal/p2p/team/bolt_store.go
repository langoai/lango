package team

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	bolt "go.etcd.io/bbolt"
)

var teamsBucket = []byte("teams")

// TeamStore persists team state.
type TeamStore interface {
	Save(team *Team) error
	Load(teamID string) (*Team, error)
	LoadAll() ([]*Team, error)
	Delete(teamID string) error
}

// BoltStore is a BoltDB-backed TeamStore.
type BoltStore struct {
	db     *bolt.DB
	logger *zap.SugaredLogger
}

// NewBoltStore creates a BoltStore and ensures the teams bucket exists.
func NewBoltStore(db *bolt.DB, logger *zap.SugaredLogger) (*BoltStore, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(teamsBucket)
		return err
	}); err != nil {
		return nil, fmt.Errorf("create teams bucket: %w", err)
	}
	return &BoltStore{db: db, logger: logger}, nil
}

// Save persists a team to BoltDB.
func (s *BoltStore) Save(t *Team) error {
	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshal team %s: %w", t.ID, err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(teamsBucket).Put([]byte(t.ID), data)
	})
}

// Load retrieves a team by ID from BoltDB.
func (s *BoltStore) Load(teamID string) (*Team, error) {
	var t Team
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(teamsBucket).Get([]byte(teamID))
		if data == nil {
			return ErrTeamNotFound
		}
		return json.Unmarshal(data, &t)
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// LoadAll retrieves all persisted teams.
func (s *BoltStore) LoadAll() ([]*Team, error) {
	var teams []*Team
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(teamsBucket)
		return b.ForEach(func(k, v []byte) error {
			var t Team
			if err := json.Unmarshal(v, &t); err != nil {
				s.logger.Warnw("skip corrupt team entry", "key", string(k), "error", err)
				return nil
			}
			teams = append(teams, &t)
			return nil
		})
	})
	return teams, err
}

// Delete removes a team from BoltDB.
func (s *BoltStore) Delete(teamID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(teamsBucket).Delete([]byte(teamID))
	})
}
