package team

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func openTestDB(t *testing.T) *bolt.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := bolt.Open(filepath.Join(dir, "test.db"), 0o600, nil)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestBoltStore_SaveAndLoad(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	store, err := NewBoltStore(db, testLogger())
	require.NoError(t, err)

	team := NewTeam("t1", "test-team", "test goal", "did:leader", 5)
	team.Budget = 100.0
	team.Spent = 25.5
	require.NoError(t, team.AddMember(&Member{
		DID:    "did:worker1",
		Name:   "worker-1",
		PeerID: "peer-w1",
		Role:   RoleWorker,
	}))

	require.NoError(t, store.Save(team))

	loaded, err := store.Load("t1")
	require.NoError(t, err)

	assert.Equal(t, team.ID, loaded.ID)
	assert.Equal(t, team.Name, loaded.Name)
	assert.Equal(t, team.Goal, loaded.Goal)
	assert.Equal(t, team.LeaderDID, loaded.LeaderDID)
	assert.Equal(t, team.Status, loaded.Status)
	assert.Equal(t, team.Budget, loaded.Budget)
	assert.Equal(t, team.Spent, loaded.Spent)
	assert.Equal(t, team.MaxMembers, loaded.MaxMembers)

	member := loaded.GetMember("did:worker1")
	require.NotNil(t, member)
	assert.Equal(t, "worker-1", member.Name)
	assert.Equal(t, RoleWorker, member.Role)
}

func TestBoltStore_LoadNotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	store, err := NewBoltStore(db, testLogger())
	require.NoError(t, err)

	_, err = store.Load("nonexistent")
	assert.ErrorIs(t, err, ErrTeamNotFound)
}

func TestBoltStore_LoadAll(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	store, err := NewBoltStore(db, testLogger())
	require.NoError(t, err)

	for _, id := range []string{"t1", "t2", "t3"} {
		team := NewTeam(id, "team-"+id, "goal", "did:leader", 3)
		require.NoError(t, store.Save(team))
	}

	teams, err := store.LoadAll()
	require.NoError(t, err)
	assert.Len(t, teams, 3)
}

func TestBoltStore_Delete(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	store, err := NewBoltStore(db, testLogger())
	require.NoError(t, err)

	team := NewTeam("t1", "test-team", "goal", "did:leader", 3)
	require.NoError(t, store.Save(team))

	require.NoError(t, store.Delete("t1"))

	_, err = store.Load("t1")
	assert.ErrorIs(t, err, ErrTeamNotFound)
}

func TestBoltStore_DeleteNonExistent(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	store, err := NewBoltStore(db, testLogger())
	require.NoError(t, err)

	// BoltDB Delete is a no-op for missing keys, so no error expected.
	assert.NoError(t, store.Delete("nonexistent"))
}

func TestTeam_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := NewTeam("t1", "team-1", "goal", "did:leader", 5)
	original.Budget = 200.0
	original.Spent = 50.0
	original.Activate()
	require.NoError(t, original.AddMember(&Member{
		DID:          "did:w1",
		Name:         "worker-1",
		PeerID:       "peer-w1",
		Role:         RoleWorker,
		Capabilities: []string{"search", "code"},
		Metadata:     map[string]string{"key": "value"},
	}))
	require.NoError(t, original.AddMember(&Member{
		DID:    "did:w2",
		Name:   "worker-2",
		PeerID: "peer-w2",
		Role:   RoleReviewer,
	}))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Team
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Status, restored.Status)
	assert.Equal(t, original.Budget, restored.Budget)
	assert.Equal(t, original.Spent, restored.Spent)
	assert.Equal(t, original.MemberCount(), restored.MemberCount())

	w1 := restored.GetMember("did:w1")
	require.NotNil(t, w1)
	assert.Equal(t, "worker-1", w1.Name)
	assert.Equal(t, RoleWorker, w1.Role)
	assert.Equal(t, []string{"search", "code"}, w1.Capabilities)
	assert.Equal(t, "value", w1.Metadata["key"])
}

func TestBoltStore_PersistAcrossReopen(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persist.db")

	// First open: create and save a team.
	db1, err := bolt.Open(dbPath, 0o600, nil)
	require.NoError(t, err)
	store1, err := NewBoltStore(db1, testLogger())
	require.NoError(t, err)

	team := NewTeam("t1", "persist-team", "goal", "did:leader", 3)
	team.Activate()
	require.NoError(t, team.AddMember(&Member{
		DID: "did:w1", Name: "worker", PeerID: "peer", Role: RoleWorker,
	}))
	require.NoError(t, store1.Save(team))
	require.NoError(t, db1.Close())

	// Second open: verify data survives.
	db2, err := bolt.Open(dbPath, 0o600, nil)
	require.NoError(t, err)
	defer db2.Close()
	store2, err := NewBoltStore(db2, testLogger())
	require.NoError(t, err)

	loaded, err := store2.Load("t1")
	require.NoError(t, err)
	assert.Equal(t, "persist-team", loaded.Name)
	assert.Equal(t, StatusActive, loaded.Status)
	assert.Equal(t, 1, loaded.MemberCount())

	// Cleanup temp file.
	_ = os.Remove(dbPath)
}

func TestCoordinator_LoadPersistedTeams(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	store, err := NewBoltStore(db, testLogger())
	require.NoError(t, err)

	// Seed store with teams in various states.
	active := NewTeam("t-active", "active-team", "goal", "did:leader", 3)
	active.Activate()
	require.NoError(t, store.Save(active))

	disbanded := NewTeam("t-disbanded", "disbanded-team", "goal", "did:leader", 3)
	disbanded.Disband()
	require.NoError(t, store.Save(disbanded))

	forming := NewTeam("t-forming", "forming-team", "goal", "did:leader", 3)
	require.NoError(t, store.Save(forming))

	// Create coordinator with the store and load.
	coord := NewCoordinator(CoordinatorConfig{
		Store:  store,
		Logger: testLogger(),
	})
	require.NoError(t, coord.LoadPersistedTeams())

	// Only active and forming teams should be loaded.
	teams := coord.ListTeams()
	assert.Len(t, teams, 2)

	_, err = coord.GetTeam("t-active")
	assert.NoError(t, err)
	_, err = coord.GetTeam("t-forming")
	assert.NoError(t, err)
	_, err = coord.GetTeam("t-disbanded")
	assert.ErrorIs(t, err, ErrTeamNotFound)
}

func TestTeam_MarshalUnmarshalPreservesTime(t *testing.T) {
	t.Parallel()

	team := NewTeam("t1", "team", "goal", "did:leader", 3)
	team.CreatedAt = time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)

	data, err := json.Marshal(team)
	require.NoError(t, err)

	var restored Team
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, team.CreatedAt.Equal(restored.CreatedAt))
}
