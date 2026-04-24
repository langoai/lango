package cockpit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/turnrunner"
)

// Deps holds the dependencies for the cockpit TUI.
// ApprovalProvider is NOT included — type assertion for SetTTYFallback
// is handled in cmd/lango/main.go's runCockpit().
type Deps struct {
	TurnRunner        *turnrunner.Runner
	Config            *config.Config
	SessionKey        string
	SessionStore      session.Store // optional; enables /mode to persist session mode
	ToolCatalog       *toolcatalog.Catalog
	MetricsCollector  *observability.MetricsCollector
	FeatureStatuses   *app.StatusCollector
	ConfigStore       storage.ConfigProfileStore
	ProfileName       string
	BackgroundManager *background.Manager    // optional, nil when unavailable
	EventBus          *eventbus.Bus          // optional, enables channel event subscription
	ApprovalHistory   *approval.HistoryStore // optional, approval decision history
	GrantStore        *approval.GrantStore   // optional, persistent session grants
}

type DeadLetterToolBridge struct {
	catalog *toolcatalog.Catalog
}

type DeadLetterListOptions struct {
	Query               string
	Adjudication        string
	LatestStatusSubtype string
}

func NewDeadLetterToolBridge(catalog *toolcatalog.Catalog) *DeadLetterToolBridge {
	return &DeadLetterToolBridge{catalog: catalog}
}

func (b *DeadLetterToolBridge) Ready() bool {
	if b == nil || b.catalog == nil {
		return false
	}
	_, hasList := b.catalog.Get("list_dead_lettered_post_adjudication_executions")
	_, hasDetail := b.catalog.Get("get_post_adjudication_execution_status")
	return hasList && hasDetail
}

func (b *DeadLetterToolBridge) CanRetry() bool {
	if b == nil || b.catalog == nil {
		return false
	}
	_, hasRetry := b.catalog.Get("retry_post_adjudication_execution")
	return hasRetry
}

func (b *DeadLetterToolBridge) List(ctx context.Context, opts DeadLetterListOptions) ([]postadjudicationstatus.DeadLetterBacklogEntry, error) {
	if b == nil || b.catalog == nil {
		return nil, fmt.Errorf("dead-letter tool catalog is not configured")
	}
	entry, ok := b.catalog.Get("list_dead_lettered_post_adjudication_executions")
	if !ok || entry.Tool == nil || entry.Tool.Handler == nil {
		return nil, fmt.Errorf("dead-letter backlog tool is not available")
	}
	params := map[string]interface{}{}
	if query := strings.TrimSpace(opts.Query); query != "" {
		params["query"] = query
	}
	switch strings.TrimSpace(opts.Adjudication) {
	case "release", "refund":
		params["adjudication"] = strings.TrimSpace(opts.Adjudication)
	}
	switch strings.TrimSpace(opts.LatestStatusSubtype) {
	case "retry-scheduled", "manual-retry-requested", "dead-lettered":
		params["latest_status_subtype"] = strings.TrimSpace(opts.LatestStatusSubtype)
	}
	raw, err := entry.Tool.Handler(ctx, params)
	if err != nil {
		return nil, err
	}
	payload, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("dead-letter backlog tool returned invalid payload")
	}
	entriesRaw, ok := payload["entries"]
	if !ok {
		return nil, fmt.Errorf("dead-letter backlog tool returned no entries")
	}
	data, err := json.Marshal(entriesRaw)
	if err != nil {
		return nil, err
	}
	var entries []postadjudicationstatus.DeadLetterBacklogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (b *DeadLetterToolBridge) Detail(ctx context.Context, transactionReceiptID string) (postadjudicationstatus.TransactionStatus, error) {
	if b == nil || b.catalog == nil {
		return postadjudicationstatus.TransactionStatus{}, fmt.Errorf("dead-letter tool catalog is not configured")
	}
	entry, ok := b.catalog.Get("get_post_adjudication_execution_status")
	if !ok || entry.Tool == nil || entry.Tool.Handler == nil {
		return postadjudicationstatus.TransactionStatus{}, fmt.Errorf("dead-letter detail tool is not available")
	}
	raw, err := entry.Tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": transactionReceiptID,
	})
	if err != nil {
		return postadjudicationstatus.TransactionStatus{}, err
	}
	status, ok := raw.(postadjudicationstatus.TransactionStatus)
	if ok {
		return status, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return postadjudicationstatus.TransactionStatus{}, err
	}
	var statusDecoded postadjudicationstatus.TransactionStatus
	if err := json.Unmarshal(data, &statusDecoded); err != nil {
		return postadjudicationstatus.TransactionStatus{}, err
	}
	return statusDecoded, nil
}

func (b *DeadLetterToolBridge) Retry(ctx context.Context, transactionReceiptID string) error {
	if b == nil || b.catalog == nil {
		return fmt.Errorf("dead-letter tool catalog is not configured")
	}
	entry, ok := b.catalog.Get("retry_post_adjudication_execution")
	if !ok || entry.Tool == nil || entry.Tool.Handler == nil {
		return fmt.Errorf("dead-letter retry tool is not available")
	}
	_, err := entry.Tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": transactionReceiptID,
	})
	return err
}
