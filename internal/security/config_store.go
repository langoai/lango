package security

import (
	"database/sql"
	"fmt"
)

// SecurityConfigDefault is the canonical row name for the primary salt/checksum
// pair used by bootstrap. Other callers may store additional rows under
// different names (e.g. per-context salts in session store tests).
const SecurityConfigDefault = "default"

// SecurityConfigStore is the single access point for the raw `security_config`
// table. It replaces the scattered `ensureSecurityTable`/`storeSalt`/`storeChecksum`
// helpers that previously lived in both `bootstrap` and `session/ent_store`.
//
// Two calling styles are supported:
//   - Default row (bootstrap): LoadSalt / StoreSalt / LoadChecksum / StoreChecksum
//     operate on the row named SecurityConfigDefault.
//   - Arbitrary name (session test suite, future multi-context): the same
//     operations are exposed with a `Named` suffix accepting a row name.
type SecurityConfigStore struct {
	db *sql.DB
}

// NewSecurityConfigStore wraps the given *sql.DB.
// EnsureTable must be called before any read/write methods.
func NewSecurityConfigStore(db *sql.DB) *SecurityConfigStore {
	return &SecurityConfigStore{db: db}
}

// EnsureTable creates the security_config table if it does not exist and adds
// the checksum column on older installations that lacked it.
func (s *SecurityConfigStore) EnsureTable() error {
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS security_config (
			name TEXT PRIMARY KEY,
			value BLOB NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create security_config table: %w", err)
	}
	var count int
	err := s.db.QueryRow(
		`SELECT count(*) FROM pragma_table_info('security_config') WHERE name='checksum'`,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check checksum column: %w", err)
	}
	if count == 0 {
		if _, err := s.db.Exec(`ALTER TABLE security_config ADD COLUMN checksum BLOB`); err != nil {
			return fmt.Errorf("add checksum column: %w", err)
		}
	}
	return nil
}

// LoadSaltNamed returns the stored salt for the named row, or nil if absent.
func (s *SecurityConfigStore) LoadSaltNamed(name string) ([]byte, error) {
	if err := s.EnsureTable(); err != nil {
		return nil, err
	}
	var salt []byte
	err := s.db.QueryRow(
		`SELECT value FROM security_config WHERE name = ?`, name,
	).Scan(&salt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query salt: %w", err)
	}
	return salt, nil
}

// StoreSaltNamed upserts the named row's salt value.
func (s *SecurityConfigStore) StoreSaltNamed(name string, salt []byte) error {
	if err := s.EnsureTable(); err != nil {
		return err
	}
	_, err := s.db.Exec(
		`INSERT INTO security_config (name, value) VALUES (?, ?)
		 ON CONFLICT(name) DO UPDATE SET value=excluded.value`,
		name, salt,
	)
	if err != nil {
		return fmt.Errorf("store salt: %w", err)
	}
	return nil
}

// LoadChecksumNamed returns the stored HMAC-SHA256 checksum for the named row, or nil.
func (s *SecurityConfigStore) LoadChecksumNamed(name string) ([]byte, error) {
	if err := s.EnsureTable(); err != nil {
		return nil, err
	}
	var raw any
	err := s.db.QueryRow(
		`SELECT checksum FROM security_config WHERE name = ?`, name,
	).Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query checksum: %w", err)
	}
	if raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unexpected checksum type %T", raw)
	}
}

// StoreChecksumNamed updates the named row's checksum column.
// The row must already exist (StoreSaltNamed before StoreChecksumNamed).
// Returns the number of rows affected so callers can detect missing-row cases.
func (s *SecurityConfigStore) StoreChecksumNamed(name string, checksum []byte) (int64, error) {
	if err := s.EnsureTable(); err != nil {
		return 0, err
	}
	res, err := s.db.Exec(
		`UPDATE security_config SET checksum = ? WHERE name = ?`,
		checksum, name,
	)
	if err != nil {
		return 0, fmt.Errorf("store checksum: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// LoadSalt returns the stored salt for the default row, or nil if absent.
func (s *SecurityConfigStore) LoadSalt() ([]byte, error) {
	return s.LoadSaltNamed(SecurityConfigDefault)
}

// StoreSalt upserts the default row's salt value.
func (s *SecurityConfigStore) StoreSalt(salt []byte) error {
	return s.StoreSaltNamed(SecurityConfigDefault, salt)
}

// LoadChecksum returns the stored HMAC-SHA256 checksum for the default row.
func (s *SecurityConfigStore) LoadChecksum() ([]byte, error) {
	return s.LoadChecksumNamed(SecurityConfigDefault)
}

// StoreChecksum updates the default row's checksum column. The row must
// already exist.
func (s *SecurityConfigStore) StoreChecksum(checksum []byte) error {
	_, err := s.StoreChecksumNamed(SecurityConfigDefault, checksum)
	return err
}

// IsFirstRun reports whether the default row is missing (no salt stored yet).
func (s *SecurityConfigStore) IsFirstRun() (bool, error) {
	salt, err := s.LoadSalt()
	if err != nil {
		return false, err
	}
	return salt == nil, nil
}
