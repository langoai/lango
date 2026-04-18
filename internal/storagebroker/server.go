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

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/dbopen"
	"github.com/langoai/lango/internal/ent"
	sec "github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
)

// Server owns the broker process state.
type Server struct {
	mu             sync.Mutex
	client         *ent.Client
	rawDB          *sql.DB
	masterKey      []byte
	payloadKey     []byte
	payloadVersion int
	sessionStore   *session.EntStore
	recallIndex    *session.RecallIndex
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
	case methodConfigLoad:
		var payload ConfigLoadRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.configLoad(ctx, payload)
	case methodConfigLoadActive:
		return s.configLoadActive(ctx)
	case methodConfigSave:
		var payload ConfigSaveRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.configSave(ctx, payload)
	case methodConfigSetActive:
		var payload ConfigSetActiveRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.configSetActive(ctx, payload)
	case methodConfigList:
		return s.configList(ctx)
	case methodConfigDelete:
		var payload ConfigDeleteRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return nil, s.configDelete(ctx, payload)
	case methodConfigExists:
		var payload ConfigExistsRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.configExists(ctx, payload)
	case methodSessionCreate:
		var payload SessionCreateRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return nil, s.sessionCreate(payload)
	case methodSessionGet:
		var payload SessionGetRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.sessionGet(payload)
	case methodSessionUpdate:
		var payload SessionUpdateRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return nil, s.sessionUpdate(payload)
	case methodSessionDelete:
		var payload SessionDeleteRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return nil, s.sessionDelete(payload)
	case methodSessionAppend:
		var payload SessionAppendMessageRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return nil, s.sessionAppend(payload)
	case methodSessionEnd:
		var payload SessionEndRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return nil, s.sessionEnd(payload)
	case methodSessionList:
		return s.sessionList()
	case methodSessionGetSalt:
		var payload SessionGetSaltRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.sessionGetSalt(payload)
	case methodSessionSetSalt:
		var payload SessionSetSaltRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return nil, s.sessionSetSalt(payload)
	case methodRecallIndex:
		var payload RecallIndexRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return nil, s.recallIndexSession(ctx, payload)
	case methodRecallProcess:
		return nil, s.recallProcessPending(ctx)
	case methodRecallSearch:
		var payload RecallSearchRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.recallSearch(ctx, payload)
	case methodRecallSummary:
		var payload RecallSummaryRequest
		if err := decodePayload(req.Payload, &payload); err != nil {
			return nil, err
		}
		return s.recallSummary(ctx, payload)
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
	s.masterKey = append([]byte(nil), req.MasterKey...)
	s.payloadKey = append([]byte(nil), req.PayloadKey...)
	s.payloadVersion = req.PayloadVersion
	s.sessionStore = session.NewEntStoreWithClient(client, session.WithDB(rawDB))
	if len(s.payloadKey) > 0 {
		s.sessionStore.SetPayloadProtector(&serverPayloadProtector{
			key:     s.payloadKey,
			version: s.payloadVersion,
		})
	}
	s.recallIndex = session.NewRecallIndex(s.sessionStore)
	if err := s.recallIndex.EnsureReady(); err != nil {
		return OpenDBResult{}, err
	}
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

func (s *Server) configStore() (*configstore.Store, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return nil, fmt.Errorf("database not opened")
	}
	crypto := sec.NewLocalCryptoProvider()
	if len(s.masterKey) == 0 {
		return nil, fmt.Errorf("master key not initialized")
	}
	if err := crypto.InitializeWithEnvelope(s.masterKey, nil); err != nil {
		return nil, err
	}
	return configstore.NewStore(s.client, crypto), nil
}

func (s *Server) configLoad(ctx context.Context, req ConfigLoadRequest) (ConfigLoadResult, error) {
	store, err := s.configStore()
	if err != nil {
		return ConfigLoadResult{}, err
	}
	cfg, keys, err := store.Load(ctx, req.Name)
	if err != nil {
		return ConfigLoadResult{}, err
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return ConfigLoadResult{}, err
	}
	return ConfigLoadResult{Config: raw, ExplicitKeys: keys}, nil
}

func (s *Server) configLoadActive(ctx context.Context) (ConfigLoadActiveResult, error) {
	store, err := s.configStore()
	if err != nil {
		return ConfigLoadActiveResult{}, err
	}
	name, cfg, keys, err := store.LoadActive(ctx)
	if err != nil {
		return ConfigLoadActiveResult{}, err
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return ConfigLoadActiveResult{}, err
	}
	return ConfigLoadActiveResult{Name: name, Config: raw, ExplicitKeys: keys}, nil
}

func (s *Server) configSave(ctx context.Context, req ConfigSaveRequest) (map[string]bool, error) {
	store, err := s.configStore()
	if err != nil {
		return nil, err
	}
	var cfg config.Config
	if err := json.Unmarshal(req.Config, &cfg); err != nil {
		return nil, fmt.Errorf("decode config payload: %w", err)
	}
	if err := store.Save(ctx, req.Name, &cfg, req.ExplicitKeys); err != nil {
		return nil, err
	}
	return map[string]bool{"saved": true}, nil
}

func (s *Server) configSetActive(ctx context.Context, req ConfigSetActiveRequest) (map[string]bool, error) {
	store, err := s.configStore()
	if err != nil {
		return nil, err
	}
	if err := store.SetActive(ctx, req.Name); err != nil {
		return nil, err
	}
	return map[string]bool{"active": true}, nil
}

func (s *Server) configList(ctx context.Context) (ConfigListResult, error) {
	store, err := s.configStore()
	if err != nil {
		return ConfigListResult{}, err
	}
	profiles, err := store.List(ctx)
	if err != nil {
		return ConfigListResult{}, err
	}
	result := ConfigListResult{Profiles: make([]ConfigProfileInfo, 0, len(profiles))}
	for _, p := range profiles {
		result.Profiles = append(result.Profiles, ConfigProfileInfo{
			Name:      p.Name,
			Active:    p.Active,
			Version:   p.Version,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
			UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (s *Server) configDelete(ctx context.Context, req ConfigDeleteRequest) error {
	store, err := s.configStore()
	if err != nil {
		return err
	}
	return store.Delete(ctx, req.Name)
}

func (s *Server) configExists(ctx context.Context, req ConfigExistsRequest) (ConfigExistsResult, error) {
	store, err := s.configStore()
	if err != nil {
		return ConfigExistsResult{}, err
	}
	exists, err := store.Exists(ctx, req.Name)
	if err != nil {
		return ConfigExistsResult{}, err
	}
	return ConfigExistsResult{Exists: exists}, nil
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
	s.sessionStore = nil
	s.recallIndex = nil
	if s.masterKey != nil {
		sec.ZeroBytes(s.masterKey)
		s.masterKey = nil
	}
	if s.payloadKey != nil {
		sec.ZeroBytes(s.payloadKey)
		s.payloadKey = nil
	}
	s.payloadVersion = 0
	return ShutdownResult{ShuttingDown: true}
}

func (s *Server) requireSessionStore() (*session.EntStore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessionStore == nil {
		return nil, fmt.Errorf("session store not initialized")
	}
	return s.sessionStore, nil
}

func (s *Server) requireRecallIndex() (*session.RecallIndex, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.recallIndex == nil {
		return nil, fmt.Errorf("recall index not initialized")
	}
	return s.recallIndex, nil
}

func (s *Server) sessionCreate(req SessionCreateRequest) error {
	store, err := s.requireSessionStore()
	if err != nil {
		return err
	}
	return store.Create(&req.Session)
}

func (s *Server) sessionGet(req SessionGetRequest) (SessionGetResult, error) {
	store, err := s.requireSessionStore()
	if err != nil {
		return SessionGetResult{}, err
	}
	sess, err := store.Get(req.Key)
	if err != nil {
		return SessionGetResult{}, err
	}
	return SessionGetResult{Session: sess}, nil
}

func (s *Server) sessionUpdate(req SessionUpdateRequest) error {
	store, err := s.requireSessionStore()
	if err != nil {
		return err
	}
	return store.Update(&req.Session)
}

func (s *Server) sessionDelete(req SessionDeleteRequest) error {
	store, err := s.requireSessionStore()
	if err != nil {
		return err
	}
	return store.Delete(req.Key)
}

func (s *Server) sessionAppend(req SessionAppendMessageRequest) error {
	store, err := s.requireSessionStore()
	if err != nil {
		return err
	}
	return store.AppendMessage(req.Key, req.Message)
}

func (s *Server) sessionEnd(req SessionEndRequest) error {
	store, err := s.requireSessionStore()
	if err != nil {
		return err
	}
	return store.End(req.Key)
}

func (s *Server) sessionList() (SessionListResult, error) {
	store, err := s.requireSessionStore()
	if err != nil {
		return SessionListResult{}, err
	}
	rows, err := store.ListSessions(context.Background())
	if err != nil {
		return SessionListResult{}, err
	}
	out := make([]SessionSummaryRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, SessionSummaryRecord{Key: row.Key, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return SessionListResult{Sessions: out}, nil
}

func (s *Server) sessionGetSalt(req SessionGetSaltRequest) (SessionGetSaltResult, error) {
	store, err := s.requireSessionStore()
	if err != nil {
		return SessionGetSaltResult{}, err
	}
	salt, err := store.GetSalt(req.Name)
	if err != nil {
		return SessionGetSaltResult{}, err
	}
	return SessionGetSaltResult{Salt: salt}, nil
}

func (s *Server) sessionSetSalt(req SessionSetSaltRequest) error {
	store, err := s.requireSessionStore()
	if err != nil {
		return err
	}
	return store.SetSalt(req.Name, req.Salt)
}

func (s *Server) recallIndexSession(ctx context.Context, req RecallIndexRequest) error {
	idx, err := s.requireRecallIndex()
	if err != nil {
		return err
	}
	return idx.IndexSession(ctx, req.Key)
}

func (s *Server) recallProcessPending(ctx context.Context) error {
	idx, err := s.requireRecallIndex()
	if err != nil {
		return err
	}
	return idx.ProcessPending(ctx)
}

func (s *Server) recallSearch(ctx context.Context, req RecallSearchRequest) (RecallSearchResult, error) {
	idx, err := s.requireRecallIndex()
	if err != nil {
		return RecallSearchResult{}, err
	}
	results, err := idx.Search(ctx, req.Query, req.Limit)
	if err != nil {
		return RecallSearchResult{}, err
	}
	out := make([]RecallSearchRecord, 0, len(results))
	for _, row := range results {
		out = append(out, RecallSearchRecord{RowID: row.RowID, Rank: row.Rank})
	}
	return RecallSearchResult{Results: out}, nil
}

func (s *Server) recallSummary(ctx context.Context, req RecallSummaryRequest) (RecallSummaryResult, error) {
	idx, err := s.requireRecallIndex()
	if err != nil {
		return RecallSummaryResult{}, err
	}
	summary, err := idx.GetSummary(ctx, req.Key)
	if err != nil {
		return RecallSummaryResult{}, err
	}
	return RecallSummaryResult{Summary: summary}, nil
}

type serverPayloadProtector struct {
	key     []byte
	version int
}

func (p *serverPayloadProtector) EncryptPayload(plaintext []byte) ([]byte, []byte, int, error) {
	ciphertext, nonce, err := sec.EncryptPayloadWithKey(p.key, plaintext)
	if err != nil {
		return nil, nil, 0, err
	}
	version := p.version
	if version == 0 {
		version = sec.PayloadKeyVersionV1
	}
	return ciphertext, nonce, version, nil
}

func (p *serverPayloadProtector) DecryptPayload(ciphertext, nonce []byte, keyVersion int) ([]byte, error) {
	if keyVersion != 0 && p.version != 0 && keyVersion != p.version {
		return nil, fmt.Errorf("unsupported payload key version %d", keyVersion)
	}
	return sec.DecryptPayloadWithKey(p.key, ciphertext, nonce)
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
