package storagebroker

import "encoding/json"

const (
	methodHealth            = "health"
	methodDBStatus          = "db_status_summary"
	methodOpenDB            = "open_db"
	methodLoadSecurityState = "load_security_state"
	methodStoreSalt         = "store_salt"
	methodStoreChecksum     = "store_checksum"
	methodShutdown          = "shutdown"
)

// Request is the persistent stdio JSON envelope sent to the broker.
type Request struct {
	ID         uint64          `json:"id"`
	Method     string          `json:"method"`
	DeadlineMS int64           `json:"deadline_ms,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

// Response is the persistent stdio JSON envelope returned by the broker.
type Response struct {
	ID     uint64          `json:"id"`
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type HealthResult struct {
	Opened bool `json:"opened"`
}

type OpenDBRequest struct {
	DBPath         string `json:"db_path"`
	EncryptionKey  string `json:"encryption_key,omitempty"`
	RawKey         bool   `json:"raw_key,omitempty"`
	CipherPageSize int    `json:"cipher_page_size,omitempty"`
}

type OpenDBResult struct {
	Opened bool `json:"opened"`
}

type DBStatusSummaryRequest struct {
	DBPath         string `json:"db_path"`
	EncryptionKey  string `json:"encryption_key,omitempty"`
	RawKey         bool   `json:"raw_key,omitempty"`
	CipherPageSize int    `json:"cipher_page_size,omitempty"`
}

type DBStatusSummaryResult struct {
	Available      bool `json:"available"`
	EncryptionKeys int  `json:"encryption_keys"`
	StoredSecrets  int  `json:"stored_secrets"`
}

type LoadSecurityStateRequest struct{}

type LoadSecurityStateResult struct {
	Salt     []byte `json:"salt,omitempty"`
	Checksum []byte `json:"checksum,omitempty"`
	FirstRun bool   `json:"first_run"`
}

type StoreSaltRequest struct {
	Salt []byte `json:"salt"`
}

type StoreChecksumRequest struct {
	Checksum []byte `json:"checksum"`
}

type ShutdownResult struct {
	ShuttingDown bool `json:"shutting_down"`
}
