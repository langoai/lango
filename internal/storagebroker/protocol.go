package storagebroker

import (
	"encoding/json"
	"time"

	"github.com/langoai/lango/internal/session"
)

const (
	methodHealth            = "health"
	methodDBStatus          = "db_status_summary"
	methodOpenDB            = "open_db"
	methodEncryptPayload    = "encrypt_payload"
	methodDecryptPayload    = "decrypt_payload"
	methodLoadSecurityState = "load_security_state"
	methodStoreSalt         = "store_salt"
	methodStoreChecksum     = "store_checksum"
	methodConfigLoad        = "config_load"
	methodConfigLoadActive  = "config_load_active"
	methodConfigSave        = "config_save"
	methodConfigSetActive   = "config_set_active"
	methodConfigList        = "config_list"
	methodConfigDelete      = "config_delete"
	methodConfigExists      = "config_exists"
	methodSessionCreate     = "session_create"
	methodSessionGet        = "session_get"
	methodSessionUpdate     = "session_update"
	methodSessionDelete     = "session_delete"
	methodSessionAppend     = "session_append_message"
	methodSessionEnd        = "session_end"
	methodSessionList       = "session_list"
	methodSessionGetSalt    = "session_get_salt"
	methodSessionSetSalt    = "session_set_salt"
	methodRecallIndex       = "recall_index_session"
	methodRecallProcess     = "recall_process_pending"
	methodRecallSearch      = "recall_search"
	methodRecallSummary     = "recall_get_summary"
	methodLearningHistory   = "learning_history"
	methodPendingInquiries  = "pending_inquiries"
	methodWorkflowRuns      = "workflow_runs"
	methodAlerts            = "alerts"
	methodReputationGet     = "reputation_get"
	methodPaymentHistory    = "payment_history"
	methodPaymentUsage      = "payment_usage"
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
	MasterKey      []byte `json:"master_key,omitempty"`
	PayloadKey     []byte `json:"payload_key,omitempty"`
	PayloadVersion int    `json:"payload_version,omitempty"`
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

type EncryptPayloadRequest struct {
	Plaintext []byte `json:"plaintext"`
}

type EncryptPayloadResult struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
	KeyVersion int    `json:"key_version"`
}

type DecryptPayloadRequest struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
	KeyVersion int    `json:"key_version"`
}

type DecryptPayloadResult struct {
	Plaintext []byte `json:"plaintext"`
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

type ConfigLoadRequest struct {
	Name string `json:"name"`
}

type ConfigLoadResult struct {
	Config       []byte          `json:"config"`
	ExplicitKeys map[string]bool `json:"explicit_keys,omitempty"`
}

type ConfigLoadActiveResult struct {
	Name         string          `json:"name"`
	Config       []byte          `json:"config"`
	ExplicitKeys map[string]bool `json:"explicit_keys,omitempty"`
}

type ConfigSaveRequest struct {
	Name         string          `json:"name"`
	Config       []byte          `json:"config"`
	ExplicitKeys map[string]bool `json:"explicit_keys,omitempty"`
}

type ConfigSetActiveRequest struct {
	Name string `json:"name"`
}

type ConfigDeleteRequest struct {
	Name string `json:"name"`
}

type ConfigExistsRequest struct {
	Name string `json:"name"`
}

type ConfigExistsResult struct {
	Exists bool `json:"exists"`
}

type ConfigProfileInfo struct {
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ConfigListResult struct {
	Profiles []ConfigProfileInfo `json:"profiles"`
}

type SessionCreateRequest struct {
	Session session.Session `json:"session"`
}

type SessionGetRequest struct {
	Key string `json:"key"`
}

type SessionGetResult struct {
	Session *session.Session `json:"session,omitempty"`
}

type SessionUpdateRequest struct {
	Session session.Session `json:"session"`
}

type SessionDeleteRequest struct {
	Key string `json:"key"`
}

type SessionAppendMessageRequest struct {
	Key     string          `json:"key"`
	Message session.Message `json:"message"`
}

type SessionEndRequest struct {
	Key string `json:"key"`
}

type SessionListResult struct {
	Sessions []SessionSummaryRecord `json:"sessions"`
}

type SessionSummaryRecord struct {
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SessionGetSaltRequest struct {
	Name string `json:"name"`
}

type SessionGetSaltResult struct {
	Salt []byte `json:"salt,omitempty"`
}

type SessionSetSaltRequest struct {
	Name string `json:"name"`
	Salt []byte `json:"salt"`
}

type RecallIndexRequest struct {
	Key string `json:"key"`
}

type RecallSearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type RecallSearchResult struct {
	Results []RecallSearchRecord `json:"results"`
}

type RecallSearchRecord struct {
	RowID string  `json:"row_id"`
	Rank  float64 `json:"rank"`
}

type RecallSummaryRequest struct {
	Key string `json:"key"`
}

type RecallSummaryResult struct {
	Summary string `json:"summary"`
}

type LearningHistoryRequest struct {
	Limit int `json:"limit"`
}

type LearningHistoryResult struct {
	Entries []LearningHistoryRecord `json:"entries"`
}

type LearningHistoryRecord struct {
	ID         string    `json:"id"`
	Trigger    string    `json:"trigger"`
	Category   string    `json:"category"`
	Diagnosis  string    `json:"diagnosis"`
	Fix        string    `json:"fix"`
	Confidence float64   `json:"confidence"`
	CreatedAt  time.Time `json:"created_at"`
}

type PendingInquiriesRequest struct {
	Limit int `json:"limit"`
}

type PendingInquiriesResult struct {
	Entries []PendingInquiryRecord `json:"entries"`
}

type PendingInquiryRecord struct {
	ID       string    `json:"id"`
	Topic    string    `json:"topic"`
	Question string    `json:"question"`
	Priority string    `json:"priority"`
	Created  time.Time `json:"created"`
}

type WorkflowRunsRequest struct {
	Limit int `json:"limit"`
}

type WorkflowRunsResult struct {
	Runs []WorkflowRunRecord `json:"runs"`
}

type WorkflowRunRecord struct {
	RunID          string    `json:"run_id"`
	WorkflowName   string    `json:"workflow_name"`
	Status         string    `json:"status"`
	TotalSteps     int       `json:"total_steps"`
	CompletedSteps int       `json:"completed_steps"`
	StartedAt      time.Time `json:"started_at"`
}

type AlertsRequest struct {
	From time.Time `json:"from"`
}

type AlertsResult struct {
	Alerts []AlertRecord `json:"alerts"`
}

type AlertRecord struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Actor     string                 `json:"actor"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
}

type ReputationGetRequest struct {
	PeerDID string `json:"peer_did"`
}

type ReputationGetResult struct {
	PeerDID             string    `json:"peer_did"`
	TrustScore          float64   `json:"trust_score"`
	SuccessfulExchanges int       `json:"successful_exchanges"`
	FailedExchanges     int       `json:"failed_exchanges"`
	TimeoutCount        int       `json:"timeout_count"`
	FirstSeen           time.Time `json:"first_seen"`
	LastInteraction     time.Time `json:"last_interaction"`
	Found               bool      `json:"found"`
}

type PaymentHistoryRequest struct {
	Limit int `json:"limit"`
}

type PaymentHistoryResult struct {
	Entries []PaymentHistoryRecord `json:"entries"`
}

type PaymentHistoryRecord struct {
	TxHash        string    `json:"tx_hash,omitempty"`
	Status        string    `json:"status"`
	Amount        string    `json:"amount"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	ChainID       int64     `json:"chain_id"`
	Purpose       string    `json:"purpose,omitempty"`
	X402URL       string    `json:"x402_url,omitempty"`
	PaymentMethod string    `json:"payment_method,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type PaymentUsageResult struct {
	DailySpent string `json:"daily_spent"`
}

type ShutdownResult struct {
	ShuttingDown bool `json:"shutting_down"`
}
