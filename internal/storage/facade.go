package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/langoai/lango/internal/agentmemory"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/ent"
	entauditlog "github.com/langoai/lango/internal/ent/auditlog"
	entinquiry "github.com/langoai/lango/internal/ent/inquiry"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/observability/audit"
	"github.com/langoai/lango/internal/observability/token"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turntrace"
	"github.com/langoai/lango/internal/workflow"
	"go.uber.org/zap"
)

// ConfigProfileStore captures the configuration profile operations needed by
// CLI and bootstrap callers. Implementations may be direct DB stores or
// broker-backed adapters.
type ConfigProfileStore interface {
	Save(ctx context.Context, name string, cfg *config.Config, explicitKeys map[string]bool) error
	Load(ctx context.Context, name string) (*config.Config, map[string]bool, error)
	LoadActive(ctx context.Context) (string, *config.Config, map[string]bool, error)
	SetActive(ctx context.Context, name string) error
	List(ctx context.Context) ([]configstore.ProfileInfo, error)
	Delete(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) (bool, error)
}

// SecurityStateStore captures bootstrap-facing security state persistence.
type SecurityStateStore interface {
	LoadSalt() ([]byte, error)
	StoreSalt(salt []byte) error
	LoadChecksum() ([]byte, error)
	StoreChecksum(checksum []byte) error
	IsFirstRun() (bool, error)
}

// ProvenanceStores groups provenance-related persistence interfaces.
type ProvenanceStores struct {
	checkpoints provenance.CheckpointStore
	sessionTree provenance.SessionTreeStore
	attribution provenance.AttributionStore
	tokenUsage  provenance.TokenUsageReader
}

func (p *ProvenanceStores) Checkpoints() provenance.CheckpointStore  { return p.checkpoints }
func (p *ProvenanceStores) SessionTree() provenance.SessionTreeStore { return p.sessionTree }
func (p *ProvenanceStores) Attribution() provenance.AttributionStore { return p.attribution }
func (p *ProvenanceStores) TokenUsage() provenance.TokenUsageReader  { return p.tokenUsage }

// SecuritySummary is a lightweight DB-backed security diagnostic snapshot.
type SecuritySummary struct {
	EncryptionKeys int
	StoredSecrets  int
}

// SandboxDecisionRecord is the storage-layer view used by `lango sandbox status`.
type SandboxDecisionRecord struct {
	Timestamp  time.Time
	SessionKey string
	Target     string
	Decision   string
	Backend    string
	Reason     string
}

type LearningHistoryRecord struct {
	ID         string
	Trigger    string
	Category   string
	Diagnosis  string
	Fix        string
	Confidence float64
	CreatedAt  time.Time
}

type InquiryRecord struct {
	ID       string
	Topic    string
	Question string
	Priority string
	Created  time.Time
}

type AlertRecord struct {
	ID        string
	Type      string
	Actor     string
	Details   map[string]interface{}
	Timestamp time.Time
}

type OntologyDeps struct {
	Registry  *ontology.EntRegistry
	Conflict  *ontology.ConflictStore
	Alias     *ontology.AliasStore
	Property  *ontology.PropertyStore
	ActionLog *ontology.ActionLogStore
}

// Option customizes a storage facade.
type Option func(*Facade)

// Facade groups the storage capabilities exposed by bootstrap.
type Facade struct {
	client           *ent.Client
	rawDB            *sql.DB
	configProfiles   ConfigProfileStore
	securityState    SecurityStateStore
	openSession      func(opts ...session.StoreOption) (session.Store, error)
	keyRegistry      func() *security.KeyRegistry
	secretsStore     func(crypto security.CryptoProvider) *security.SecretsStore
	runLedger        func() runledger.RunLedgerStore
	cronStore        func() cron.Store
	turnTrace        func() turntrace.Store
	agentMemory      func() agentmemory.Store
	provenance       *ProvenanceStores
	securityDiag     func(ctx context.Context) (SecuritySummary, error)
	sandboxDecisions func(ctx context.Context, sessionPrefix string, limit int) ([]SandboxDecisionRecord, error)
	auditRecorder    func() *audit.Recorder
	tokenStore       func() *token.EntTokenStore
	learningHistory  func(ctx context.Context, limit int) ([]LearningHistoryRecord, error)
	pendingInquiries func(ctx context.Context, limit int) ([]InquiryRecord, error)
	alerts           func(ctx context.Context, from time.Time) ([]AlertRecord, error)
	ontologyDeps     func() *OntologyDeps
	reputationStore  func(logger *zap.SugaredLogger) *reputation.Store
	workflowState    func(logger *zap.SugaredLogger) workflow.RunStore
	closeFn          func() error
}

// NewFacade constructs a storage facade from capability implementations.
func NewFacade(configProfiles ConfigProfileStore, securityState SecurityStateStore, opts ...Option) *Facade {
	f := &Facade{
		configProfiles: configProfiles,
		securityState:  securityState,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(f)
		}
	}
	return f
}

func (f *Facade) ConfigProfiles() ConfigProfileStore { return f.configProfiles }
func (f *Facade) SecurityState() SecurityStateStore  { return f.securityState }

func (f *Facade) OpenSessionStore(opts ...session.StoreOption) (session.Store, error) {
	if f == nil || f.openSession == nil {
		return nil, fmt.Errorf("session storage unavailable")
	}
	return f.openSession(opts...)
}

// FTSDB exposes the shared SQLite handle only for domain-specific FTS index
// initialization while broader raw DB access remains hidden from app/CLI code.
func (f *Facade) FTSDB() *sql.DB {
	if f == nil {
		return nil
	}
	return f.rawDB
}

func (f *Facade) KeyRegistry() *security.KeyRegistry {
	if f == nil || f.keyRegistry == nil {
		return nil
	}
	return f.keyRegistry()
}

func (f *Facade) SecretsStore(crypto security.CryptoProvider) *security.SecretsStore {
	if f == nil || f.secretsStore == nil || crypto == nil {
		return nil
	}
	return f.secretsStore(crypto)
}

func (f *Facade) RunLedger() runledger.RunLedgerStore {
	if f == nil || f.runLedger == nil {
		return nil
	}
	return f.runLedger()
}

func (f *Facade) Cron() cron.Store {
	if f == nil || f.cronStore == nil {
		return nil
	}
	return f.cronStore()
}

func (f *Facade) TurnTrace() turntrace.Store {
	if f == nil || f.turnTrace == nil {
		return nil
	}
	return f.turnTrace()
}

func (f *Facade) AgentMemory() agentmemory.Store {
	if f == nil || f.agentMemory == nil {
		return nil
	}
	return f.agentMemory()
}

func (f *Facade) Provenance() *ProvenanceStores {
	if f == nil {
		return nil
	}
	return f.provenance
}

func (f *Facade) SecuritySummary(ctx context.Context) (SecuritySummary, error) {
	if f == nil || f.securityDiag == nil {
		return SecuritySummary{}, fmt.Errorf("security diagnostics unavailable")
	}
	return f.securityDiag(ctx)
}

func (f *Facade) RecentSandboxDecisions(ctx context.Context, sessionPrefix string, limit int) ([]SandboxDecisionRecord, error) {
	if f == nil || f.sandboxDecisions == nil {
		return nil, fmt.Errorf("audit storage unavailable")
	}
	return f.sandboxDecisions(ctx, sessionPrefix, limit)
}

func (f *Facade) AuditRecorder() *audit.Recorder {
	if f == nil || f.auditRecorder == nil {
		return nil
	}
	return f.auditRecorder()
}

func (f *Facade) TokenStore() *token.EntTokenStore {
	if f == nil || f.tokenStore == nil {
		return nil
	}
	return f.tokenStore()
}

func (f *Facade) LearningHistory(ctx context.Context, limit int) ([]LearningHistoryRecord, error) {
	if f == nil || f.learningHistory == nil {
		return nil, fmt.Errorf("learning storage unavailable")
	}
	return f.learningHistory(ctx, limit)
}

func (f *Facade) PendingInquiries(ctx context.Context, limit int) ([]InquiryRecord, error) {
	if f == nil || f.pendingInquiries == nil {
		return nil, fmt.Errorf("inquiry storage unavailable")
	}
	return f.pendingInquiries(ctx, limit)
}

func (f *Facade) Alerts(ctx context.Context, from time.Time) ([]AlertRecord, error) {
	if f == nil || f.alerts == nil {
		return nil, fmt.Errorf("alert storage unavailable")
	}
	return f.alerts(ctx, from)
}

func (f *Facade) OntologyDeps() *OntologyDeps {
	if f == nil || f.ontologyDeps == nil {
		return nil
	}
	return f.ontologyDeps()
}

func (f *Facade) ReputationStore(logger *zap.SugaredLogger) *reputation.Store {
	if f == nil || f.reputationStore == nil {
		return nil
	}
	return f.reputationStore(logger)
}

func (f *Facade) WorkflowStateStore(logger *zap.SugaredLogger) workflow.RunStore {
	if f == nil || f.workflowState == nil {
		return nil
	}
	return f.workflowState(logger)
}

func (f *Facade) PaymentClient() *ent.Client {
	if f == nil {
		return nil
	}
	return f.client
}

func (f *Facade) Close() error {
	if f == nil || f.closeFn == nil {
		return nil
	}
	return f.closeFn()
}

// WithSessionStoreFactory wires session store creation into the facade.
func WithSessionStoreFactory(fn func(opts ...session.StoreOption) (session.Store, error)) Option {
	return func(f *Facade) {
		f.openSession = fn
	}
}

// WithEntClient wires ent-backed storage capabilities into the facade.
func WithEntClient(client *ent.Client) Option {
	return func(f *Facade) {
		if client == nil {
			return
		}
		f.client = client

		f.keyRegistry = func() *security.KeyRegistry {
			return security.NewKeyRegistry(client)
		}
		f.secretsStore = func(crypto security.CryptoProvider) *security.SecretsStore {
			registry := security.NewKeyRegistry(client)
			return security.NewSecretsStore(client, registry, crypto)
		}
		f.runLedger = func() runledger.RunLedgerStore {
			return runledger.NewEntStore(client)
		}
		f.cronStore = func() cron.Store {
			return cron.NewEntStore(client)
		}
		f.turnTrace = func() turntrace.Store {
			return turntrace.NewEntStore(client)
		}
		f.agentMemory = func() agentmemory.Store {
			return agentmemory.NewEntStore(client)
		}
		f.provenance = &ProvenanceStores{
			checkpoints: provenance.NewEntCheckpointStore(client),
			sessionTree: provenance.NewEntSessionTreeStore(client),
			attribution: provenance.NewEntAttributionStore(client),
			tokenUsage:  token.NewEntTokenStore(client),
		}
		f.securityDiag = func(ctx context.Context) (SecuritySummary, error) {
			keys, err := client.Key.Query().Count(ctx)
			if err != nil {
				return SecuritySummary{}, err
			}
			secretsCount, err := client.Secret.Query().Count(ctx)
			if err != nil {
				return SecuritySummary{}, err
			}
			return SecuritySummary{
				EncryptionKeys: keys,
				StoredSecrets:  secretsCount,
			}, nil
		}
		f.sandboxDecisions = func(ctx context.Context, sessionPrefix string, limit int) ([]SandboxDecisionRecord, error) {
			if limit <= 0 {
				limit = 10
			}

			query := client.AuditLog.Query().
				Where(entauditlog.ActionEQ(entauditlog.ActionSandboxDecision)).
				Order(entauditlog.ByTimestamp(entsql.OrderDesc())).
				Limit(limit)
			if sessionPrefix != "" {
				query = query.Where(entauditlog.SessionKeyHasPrefix(sessionPrefix))
			}

			rows, err := query.All(ctx)
			if err != nil {
				return nil, err
			}

			out := make([]SandboxDecisionRecord, 0, len(rows))
			for _, row := range rows {
				record := SandboxDecisionRecord{
					Timestamp:  row.Timestamp,
					SessionKey: row.SessionKey,
					Target:     row.Target,
				}
				if v, ok := row.Details["decision"].(string); ok {
					record.Decision = v
				}
				if v, ok := row.Details["backend"].(string); ok {
					record.Backend = v
				}
				if v, ok := row.Details["reason"].(string); ok {
					record.Reason = v
				}
				out = append(out, record)
			}
			return out, nil
		}
		f.auditRecorder = func() *audit.Recorder {
			return audit.NewRecorder(client)
		}
		f.tokenStore = func() *token.EntTokenStore {
			return token.NewEntTokenStore(client)
		}
		f.learningHistory = func(ctx context.Context, limit int) ([]LearningHistoryRecord, error) {
			if limit <= 0 {
				limit = 20
			}
			rows, err := client.Learning.Query().
				Order(entlearning.ByCreatedAt(entsql.OrderDesc())).
				Limit(limit).
				All(ctx)
			if err != nil {
				return nil, err
			}
			out := make([]LearningHistoryRecord, 0, len(rows))
			for _, row := range rows {
				out = append(out, LearningHistoryRecord{
					ID:         row.ID.String(),
					Trigger:    row.Trigger,
					Category:   string(row.Category),
					Diagnosis:  row.Diagnosis,
					Fix:        row.Fix,
					Confidence: row.Confidence,
					CreatedAt:  row.CreatedAt,
				})
			}
			return out, nil
		}
		f.pendingInquiries = func(ctx context.Context, limit int) ([]InquiryRecord, error) {
			if limit <= 0 {
				limit = 20
			}
			rows, err := client.Inquiry.Query().
				Where(entinquiry.StatusEQ(entinquiry.StatusPending)).
				Order(entinquiry.ByCreatedAt()).
				Limit(limit).
				All(ctx)
			if err != nil {
				return nil, err
			}
			out := make([]InquiryRecord, 0, len(rows))
			for _, row := range rows {
				out = append(out, InquiryRecord{
					ID:       row.ID.String(),
					Topic:    row.Topic,
					Question: row.Question,
					Priority: string(row.Priority),
					Created:  row.CreatedAt,
				})
			}
			return out, nil
		}
		f.alerts = func(ctx context.Context, from time.Time) ([]AlertRecord, error) {
			rows, err := client.AuditLog.Query().
				Where(
					entauditlog.ActionEQ(entauditlog.Action("alert")),
					entauditlog.TimestampGTE(from),
				).
				Order(entauditlog.ByTimestamp(entsql.OrderDesc())).
				All(ctx)
			if err != nil {
				return nil, err
			}
			out := make([]AlertRecord, 0, len(rows))
			for _, row := range rows {
				out = append(out, AlertRecord{
					ID:        row.ID.String(),
					Type:      row.Target,
					Actor:     row.Actor,
					Details:   row.Details,
					Timestamp: row.Timestamp,
				})
			}
			return out, nil
		}
		f.ontologyDeps = func() *OntologyDeps {
			return &OntologyDeps{
				Registry:  ontology.NewEntRegistry(client),
				Conflict:  ontology.NewConflictStore(client),
				Alias:     ontology.NewAliasStore(client),
				Property:  ontology.NewPropertyStore(client),
				ActionLog: ontology.NewActionLogStore(client),
			}
		}
		f.reputationStore = func(logger *zap.SugaredLogger) *reputation.Store {
			return reputation.NewStore(client, logger)
		}
		f.workflowState = func(logger *zap.SugaredLogger) workflow.RunStore {
			return workflow.NewStateStore(client, logger)
		}
		f.closeFn = client.Close
	}
}

// WithRawDB exposes the shared SQL DB to transitional callers that still need
// low-level read/write access through the storage facade.
func WithRawDB(db *sql.DB) Option {
	return func(f *Facade) {
		if f == nil || db == nil {
			return
		}
		f.rawDB = db
		if f.closeFn == nil {
			f.closeFn = db.Close
		}
	}
}

// WithSessionDBPath configures the direct ent-backed session store opener used
// until the broker owns session store construction outright.
func WithSessionDBPath(path string) Option {
	return func(f *Facade) {
		if path == "" || f == nil || f.openSession != nil {
			return
		}
		f.openSession = func(opts ...session.StoreOption) (session.Store, error) {
			return session.NewEntStore(path, opts...)
		}
	}
}

// WithSessionClient configures the facade to reuse the shared ent client for
// session storage when available.
func WithSessionClient(client *ent.Client) Option {
	return func(f *Facade) {
		if client == nil || f == nil {
			return
		}
		f.openSession = func(opts ...session.StoreOption) (session.Store, error) {
			return session.NewEntStoreWithClient(client, opts...), nil
		}
	}
}

// WithKeyRegistryFactory overrides how key registries are created.
func WithKeyRegistryFactory(fn func() *security.KeyRegistry) Option {
	return func(f *Facade) {
		f.keyRegistry = fn
	}
}

// WithSecurityDiagnostics overrides DB-backed security summary reads.
func WithSecurityDiagnostics(fn func(ctx context.Context) (SecuritySummary, error)) Option {
	return func(f *Facade) {
		f.securityDiag = fn
	}
}

// WithSandboxDecisionReader overrides audit decision reads.
func WithSandboxDecisionReader(fn func(ctx context.Context, sessionPrefix string, limit int) ([]SandboxDecisionRecord, error)) Option {
	return func(f *Facade) {
		f.sandboxDecisions = fn
	}
}

// WithSecretsStoreFactory overrides secrets store construction.
func WithSecretsStoreFactory(fn func(crypto security.CryptoProvider) *security.SecretsStore) Option {
	return func(f *Facade) {
		f.secretsStore = fn
	}
}

// WithRunLedgerFactory overrides runledger store construction.
func WithRunLedgerFactory(fn func() runledger.RunLedgerStore) Option {
	return func(f *Facade) {
		f.runLedger = fn
	}
}

// WithTurnTraceFactory overrides trace store construction.
func WithTurnTraceFactory(fn func() turntrace.Store) Option {
	return func(f *Facade) {
		f.turnTrace = fn
	}
}

// WithCronFactory overrides cron store construction.
func WithCronFactory(fn func() cron.Store) Option {
	return func(f *Facade) {
		f.cronStore = fn
	}
}

// WithAgentMemoryFactory overrides agent memory store construction.
func WithAgentMemoryFactory(fn func() agentmemory.Store) Option {
	return func(f *Facade) {
		f.agentMemory = fn
	}
}

// WithProvenanceStores overrides provenance store wiring.
func WithProvenanceStores(p *ProvenanceStores) Option {
	return func(f *Facade) {
		f.provenance = p
	}
}
