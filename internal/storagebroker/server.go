package storagebroker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/langoai/lango/internal/dbopen"
	"github.com/langoai/lango/internal/ent"
	sec "github.com/langoai/lango/internal/security"
)

// Server owns the broker process state.
type Server struct {
	mu             sync.Mutex
	client         *ent.Client
	rawDB          *sql.DB
	payloadKey     []byte
	payloadVersion int
	stopped        bool
}

// NewServer constructs a broker server with no open database yet.
func NewServer() *Server {
	return &Server{}
}

// Run serves newline-delimited JSON requests until EOF or shutdown.
func (s *Server) Run(in io.Reader, out io.Writer) error {
	dec := json.NewDecoder(in)
	enc := json.NewEncoder(out)

	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("decode broker request: %w", err)
		}

		resp := s.handle(req)
		if err := enc.Encode(resp); err != nil {
			return fmt.Errorf("encode broker response: %w", err)
		}

		s.mu.Lock()
		stopped := s.stopped
		s.mu.Unlock()
		if stopped {
			return nil
		}
	}
}

func (s *Server) handle(req Request) Response {
	resp := Response{ID: req.ID}

	ctx := context.Background()
	if req.DeadlineMS > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.DeadlineMS)*time.Millisecond)
		defer cancel()
	}

	result, err := s.dispatch(ctx, req)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	resp.OK = true
	if result != nil {
		raw, err := json.Marshal(result)
		if err != nil {
			resp.OK = false
			resp.Error = fmt.Sprintf("marshal broker result: %v", err)
			return resp
		}
		resp.Result = raw
	}
	return resp
}

func (s *Server) dispatch(ctx context.Context, req Request) (interface{}, error) {
	switch req.Method {
	case methodHealth:
		return s.health(), nil
	case methodDBStatus:
		var payload DBStatusSummaryRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.dbStatus(ctx, payload)
	case methodEncryptPayload:
		var payload EncryptPayloadRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.encryptPayload(ctx, payload)
	case methodDecryptPayload:
		var payload DecryptPayloadRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.decryptPayload(ctx, payload)
	case methodLoadSecurityState:
		return s.loadSecurityState(ctx)
	case methodStoreSalt:
		var payload StoreSaltRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.storeSalt(ctx, payload)
	case methodStoreChecksum:
		var payload StoreChecksumRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.storeChecksum(ctx, payload)
	case methodOpenDB:
		var payload OpenDBRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.openDB(ctx, payload)
	case methodShutdown:
		return s.shutdown(), nil
	default:
		return nil, fmt.Errorf("unknown broker method %q", req.Method)
	}
}

func (s *Server) health() HealthResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	return HealthResult{Opened: s.client != nil && s.rawDB != nil}
}

func (s *Server) openDB(_ context.Context, req OpenDBRequest) (OpenDBResult, error) {
	if req.DBPath == "" {
		return OpenDBResult{}, fmt.Errorf("open_db requires db_path")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil || s.rawDB != nil {
		return OpenDBResult{Opened: true}, nil
	}

	client, rawDB, err := dbopen.OpenManaged(req.DBPath, req.EncryptionKey, req.RawKey, req.CipherPageSize)
	if err != nil {
		return OpenDBResult{}, err
	}

	s.client = client
	s.rawDB = rawDB
	s.payloadKey = append([]byte(nil), req.PayloadKey...)
	s.payloadVersion = req.PayloadVersion
	return OpenDBResult{Opened: true}, nil
}

func (s *Server) dbStatus(ctx context.Context, req DBStatusSummaryRequest) (DBStatusSummaryResult, error) {
	if req.DBPath == "" {
		return DBStatusSummaryResult{}, fmt.Errorf("db_status_summary requires db_path")
	}

	client, rawDB, err := dbopen.OpenReadOnly(req.DBPath, req.EncryptionKey, req.RawKey, req.CipherPageSize)
	if err != nil {
		return DBStatusSummaryResult{}, err
	}
	defer client.Close()
	defer rawDB.Close()

	result := DBStatusSummaryResult{Available: true}

	registry := sec.NewKeyRegistry(client)
	if keys, err := registry.ListKeys(ctx); err == nil {
		result.EncryptionKeys = len(keys)
	}
	if n, err := client.Secret.Query().Count(ctx); err == nil {
		result.StoredSecrets = n
	}

	return result, nil
}

func (s *Server) loadSecurityState(_ context.Context) (LoadSecurityStateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rawDB == nil {
		return LoadSecurityStateResult{}, fmt.Errorf("database not opened")
	}

	store := sec.NewSecurityConfigStore(s.rawDB)
	salt, err := store.LoadSalt()
	if err != nil {
		return LoadSecurityStateResult{}, err
	}
	checksum, err := store.LoadChecksum()
	if err != nil {
		return LoadSecurityStateResult{}, err
	}
	firstRun, err := store.IsFirstRun()
	if err != nil {
		return LoadSecurityStateResult{}, err
	}
	return LoadSecurityStateResult{
		Salt:     salt,
		Checksum: checksum,
		FirstRun: firstRun,
	}, nil
}

func (s *Server) encryptPayload(_ context.Context, req EncryptPayloadRequest) (EncryptPayloadResult, error) {
	s.mu.Lock()
	key := append([]byte(nil), s.payloadKey...)
	version := s.payloadVersion
	s.mu.Unlock()
	if len(key) == 0 {
		return EncryptPayloadResult{}, fmt.Errorf("payload protection key not initialized")
	}
	if version == 0 {
		version = sec.PayloadKeyVersionV1
	}
	ciphertext, nonce, err := sec.EncryptPayloadWithKey(key, req.Plaintext)
	if err != nil {
		return EncryptPayloadResult{}, err
	}
	return EncryptPayloadResult{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		KeyVersion: version,
	}, nil
}

func (s *Server) decryptPayload(_ context.Context, req DecryptPayloadRequest) (DecryptPayloadResult, error) {
	s.mu.Lock()
	key := append([]byte(nil), s.payloadKey...)
	version := s.payloadVersion
	s.mu.Unlock()
	if len(key) == 0 {
		return DecryptPayloadResult{}, fmt.Errorf("payload protection key not initialized")
	}
	if req.KeyVersion != 0 && version != 0 && req.KeyVersion != version {
		return DecryptPayloadResult{}, fmt.Errorf("unsupported payload key version %d", req.KeyVersion)
	}
	plaintext, err := sec.DecryptPayloadWithKey(key, req.Ciphertext, req.Nonce)
	if err != nil {
		return DecryptPayloadResult{}, err
	}
	return DecryptPayloadResult{Plaintext: plaintext}, nil
}

func (s *Server) storeSalt(_ context.Context, req StoreSaltRequest) (map[string]bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rawDB == nil {
		return nil, fmt.Errorf("database not opened")
	}
	if err := sec.NewSecurityConfigStore(s.rawDB).StoreSalt(req.Salt); err != nil {
		return nil, err
	}
	return map[string]bool{"stored": true}, nil
}

func (s *Server) storeChecksum(_ context.Context, req StoreChecksumRequest) (map[string]bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rawDB == nil {
		return nil, fmt.Errorf("database not opened")
	}
	if err := sec.NewSecurityConfigStore(s.rawDB).StoreChecksum(req.Checksum); err != nil {
		return nil, err
	}
	return map[string]bool{"stored": true}, nil
}

func (s *Server) shutdown() ShutdownResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stopped = true
	if s.client != nil {
		_ = s.client.Close()
		s.client = nil
	}
	if s.rawDB != nil {
		_ = s.rawDB.Close()
		s.rawDB = nil
	}
	if s.payloadKey != nil {
		sec.ZeroBytes(s.payloadKey)
		s.payloadKey = nil
	}
	s.payloadVersion = 0
	return ShutdownResult{ShuttingDown: true}
}

func decodePayload(data json.RawMessage, out interface{}) error {
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode broker payload: %w", err)
	}
	return nil
}
