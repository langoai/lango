package security

import (
	"bytes"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSecurityConfigStore_FirstRunAndStoreSalt(t *testing.T) {
	db := newTestDB(t)
	store := NewSecurityConfigStore(db)

	first, err := store.IsFirstRun()
	if err != nil {
		t.Fatalf("IsFirstRun: %v", err)
	}
	if !first {
		t.Fatal("expected first run on fresh db")
	}

	salt := []byte("0123456789abcdef")
	if err := store.StoreSalt(salt); err != nil {
		t.Fatalf("StoreSalt: %v", err)
	}

	got, err := store.LoadSalt()
	if err != nil {
		t.Fatalf("LoadSalt: %v", err)
	}
	if !bytes.Equal(got, salt) {
		t.Fatalf("salt mismatch: got %x want %x", got, salt)
	}

	first, err = store.IsFirstRun()
	if err != nil {
		t.Fatalf("IsFirstRun after store: %v", err)
	}
	if first {
		t.Fatal("expected not-first-run after StoreSalt")
	}
}

func TestSecurityConfigStore_ChecksumRoundTrip(t *testing.T) {
	db := newTestDB(t)
	store := NewSecurityConfigStore(db)

	if err := store.StoreSalt([]byte("saltbytes--16bts")); err != nil {
		t.Fatalf("StoreSalt: %v", err)
	}
	sum := []byte("hmac-sha256-checksum-32-bytes-ab")
	if err := store.StoreChecksum(sum); err != nil {
		t.Fatalf("StoreChecksum: %v", err)
	}
	got, err := store.LoadChecksum()
	if err != nil {
		t.Fatalf("LoadChecksum: %v", err)
	}
	if !bytes.Equal(got, sum) {
		t.Fatalf("checksum mismatch: got %x want %x", got, sum)
	}
}

func TestSecurityConfigStore_LoadSaltMissing(t *testing.T) {
	db := newTestDB(t)
	store := NewSecurityConfigStore(db)

	got, err := store.LoadSalt()
	if err != nil {
		t.Fatalf("LoadSalt on empty: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil salt, got %x", got)
	}
}

func TestSecurityConfigStore_LoadChecksumMissing(t *testing.T) {
	db := newTestDB(t)
	store := NewSecurityConfigStore(db)

	if err := store.StoreSalt([]byte("some-salt-16byte")); err != nil {
		t.Fatal(err)
	}
	got, err := store.LoadChecksum()
	if err != nil {
		t.Fatalf("LoadChecksum missing: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil checksum, got %x", got)
	}
}
