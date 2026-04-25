package pages

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/receipts"
)

type mockDeadLetterListFn struct {
	items       []postadjudicationstatus.DeadLetterBacklogEntry
	err         error
	called      int
	lastOptions DeadLetterListOptions
}

func (m *mockDeadLetterListFn) call(_ context.Context, opts DeadLetterListOptions) ([]postadjudicationstatus.DeadLetterBacklogEntry, error) {
	m.called++
	m.lastOptions = opts
	if m.err != nil {
		return nil, m.err
	}
	filtered := make([]postadjudicationstatus.DeadLetterBacklogEntry, 0, len(m.items))
	query := strings.ToLower(strings.TrimSpace(opts.Query))
	actor := strings.TrimSpace(opts.ManualReplayActor)
	reasonQuery := strings.ToLower(strings.TrimSpace(opts.DeadLetterReasonQuery))
	dispatchReference := strings.TrimSpace(opts.LatestDispatchReference)
	family := strings.TrimSpace(opts.LatestStatusSubtypeFamily)
	anyMatchFamily := strings.TrimSpace(opts.AnyMatchFamily)
	after := parseDeadLetterTime(strings.TrimSpace(opts.DeadLetteredAfter))
	before := parseDeadLetterTime(strings.TrimSpace(opts.DeadLetteredBefore))
	for _, item := range m.items {
		if adjudication := strings.TrimSpace(opts.Adjudication); adjudication != "" && !strings.EqualFold(item.Adjudication, adjudication) {
			continue
		}
		if subtype := strings.TrimSpace(opts.LatestStatusSubtype); subtype != "" && !strings.EqualFold(item.LatestStatusSubtype, subtype) {
			continue
		}
		if family != "" && !strings.EqualFold(item.LatestStatusSubtypeFamily, family) {
			continue
		}
		if anyMatchFamily != "" && !containsFamily(item.AnyMatchFamilies, anyMatchFamily) {
			continue
		}
		if actor != "" && !strings.EqualFold(item.LatestManualReplayActor, actor) {
			continue
		}
		if reasonQuery != "" && !strings.Contains(strings.ToLower(item.LatestDeadLetterReason), reasonQuery) {
			continue
		}
		if dispatchReference != "" && !strings.EqualFold(item.LatestDispatchReference, dispatchReference) {
			continue
		}
		if !after.IsZero() {
			itemTime := parseDeadLetterTime(item.LatestDeadLetteredAt)
			if itemTime.IsZero() || itemTime.Before(after) {
				continue
			}
		}
		if !before.IsZero() {
			itemTime := parseDeadLetterTime(item.LatestDeadLetteredAt)
			if itemTime.IsZero() || itemTime.After(before) {
				continue
			}
		}
		if query != "" {
			tx := strings.ToLower(item.TransactionReceiptID)
			sub := strings.ToLower(item.SubmissionReceiptID)
			if !strings.Contains(tx, query) && !strings.Contains(sub, query) {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func containsFamily(families []string, want string) bool {
	for _, family := range families {
		if strings.EqualFold(strings.TrimSpace(family), strings.TrimSpace(want)) {
			return true
		}
	}
	return false
}

func parseDeadLetterTime(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}
	return parsed
}

type mockDeadLetterDetailFn struct {
	statusByID map[string]postadjudicationstatus.TransactionStatus
	errByID    map[string]error
	calls      []string
}

func (m *mockDeadLetterDetailFn) call(_ context.Context, transactionReceiptID string) (postadjudicationstatus.TransactionStatus, error) {
	m.calls = append(m.calls, transactionReceiptID)
	if err, ok := m.errByID[transactionReceiptID]; ok {
		return postadjudicationstatus.TransactionStatus{}, err
	}
	return m.statusByID[transactionReceiptID], nil
}

type mockDeadLetterRetryFn struct {
	errByID map[string]error
	calls   []string
}

func (m *mockDeadLetterRetryFn) call(_ context.Context, transactionReceiptID string) error {
	m.calls = append(m.calls, transactionReceiptID)
	if err, ok := m.errByID[transactionReceiptID]; ok {
		return err
	}
	return nil
}

type sequenceDeadLetterListFn struct {
	results     [][]postadjudicationstatus.DeadLetterBacklogEntry
	errs        []error
	called      int
	lastOptions DeadLetterListOptions
}

func (m *sequenceDeadLetterListFn) call(_ context.Context, opts DeadLetterListOptions) ([]postadjudicationstatus.DeadLetterBacklogEntry, error) {
	m.called++
	m.lastOptions = opts

	idx := m.called - 1
	if idx >= len(m.results) {
		idx = len(m.results) - 1
	}
	if idx < 0 {
		idx = 0
	}

	if len(m.errs) > 0 {
		errIdx := m.called - 1
		if errIdx >= len(m.errs) {
			errIdx = len(m.errs) - 1
		}
		if errIdx >= 0 && m.errs[errIdx] != nil {
			return nil, m.errs[errIdx]
		}
	}

	items := m.results[idx]
	return append([]postadjudicationstatus.DeadLetterBacklogEntry(nil), items...), nil
}

func TestDeadLettersPage_Title(t *testing.T) {
	page := NewDeadLettersPage(nil, nil)
	assert.Equal(t, "Dead Letters", page.Title())
}

func TestDeadLettersPage_ShortHelp(t *testing.T) {
	page := NewDeadLettersPage(nil, nil)
	bindings := page.ShortHelp()
	require.Len(t, bindings, 10)
}

func TestDeadLettersPage_ShortHelpIncludesRetryWhenEnabled(t *testing.T) {
	page := NewDeadLettersPage(nil, nil, (&mockDeadLetterRetryFn{}).call)
	page.detail = &postadjudicationstatus.TransactionStatus{CanRetry: true}
	bindings := page.ShortHelp()
	require.Len(t, bindings, 11)
	assert.Equal(t, "r", bindings[10].Keys()[0])
}

func TestDeadLettersPage_ShortHelpShowsConfirmWhenPending(t *testing.T) {
	page := NewDeadLettersPage(nil, nil, (&mockDeadLetterRetryFn{}).call)
	page.selectedID = "tx-1"
	page.detail = &postadjudicationstatus.TransactionStatus{CanRetry: true}
	page.retryConfirmID = "tx-1"

	bindings := page.ShortHelp()
	require.Len(t, bindings, 11)
	assert.Equal(t, "confirm", bindings[10].Help().Desc)
}

func TestDeadLettersPage_ActivateLoadsBacklogAndDetail(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{
				TransactionReceiptID:   "tx-1",
				SubmissionReceiptID:    "sub-1",
				Adjudication:           string(receipts.EscrowAdjudicationRelease),
				LatestRetryAttempt:     3,
				LatestDeadLetterReason: "worker exhausted",
				IsDeadLettered:         true,
				CanRetry:               true,
			},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestDeadLetterReason:  "worker exhausted",
					LatestRetryAttempt:      3,
					LatestDispatchReference: "dispatch-1",
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   string(receipts.EscrowAdjudicationRelease),
			},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	cmd := page.Activate()
	require.NotNil(t, cmd)

	updated, detailCmd := page.Update(cmd())
	page = updated.(*DeadLettersPage)
	require.Len(t, page.items, 1)
	assert.Equal(t, "tx-1", page.selectedID)
	require.NotNil(t, detailCmd)
	assert.Equal(t, DeadLetterListOptions{}, listFn.lastOptions)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-1", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
	assert.Equal(t, []string{"tx-1"}, detailFn.calls)
}

func TestDeadLettersPage_SelectionLoadsNextDetail(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 2, LatestDeadLetterReason: "release failed", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "refund", LatestRetryAttempt: 4, LatestDeadLetterReason: "refund failed", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"}}, Adjudication: "release"},
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, Adjudication: "refund"},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	assert.Equal(t, 1, page.cursor)
	assert.Equal(t, "tx-2", page.selectedID)
	require.NotNil(t, detailCmd)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-2", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
	assert.Equal(t, []string{"tx-1", "tx-2"}, detailFn.calls)
}

func TestDeadLettersPage_ApplyFiltersReloadsAndResetsSelection(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "release", LatestStatusSubtype: "manual-retry-requested", LatestRetryAttempt: 4, LatestDeadLetterReason: "release failed", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-3", SubmissionReceiptID: "sub-3", Adjudication: "release", LatestStatusSubtype: "dead-lettered", LatestStatusSubtypeFamily: "dead-letter", AnyMatchFamilies: []string{"dead-letter"}, LatestRetryAttempt: 2, LatestDeadLetterReason: "release failed again", LatestDispatchReference: "dispatch-3", LatestManualReplayActor: "operator:bob", LatestDeadLetteredAt: "2026-04-24T10:00:00Z", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "release", LatestStatusSubtype: "manual-retry-requested", LatestStatusSubtypeFamily: "manual-retry", AnyMatchFamilies: []string{"manual-retry", "retry"}, LatestDispatchReference: "dispatch-2", LatestManualReplayActor: "operator:alice", LatestDeadLetteredAt: "2026-04-24T12:00:00Z", LatestRetryAttempt: 4, LatestDeadLetterReason: "release failed", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, Adjudication: "release"},
			"tx-3": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-3"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-3"}}, Adjudication: "release"},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyRunes, Runes: []rune("x")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("p")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("l")},
		{Type: tea.KeyRunes, Runes: []rune("i")},
		{Type: tea.KeyRunes, Runes: []rune("c")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("6")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("4")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("4")},
		{Type: tea.KeyRunes, Runes: []rune("T")},
		{Type: tea.KeyRunes, Runes: []rune("1")},
		{Type: tea.KeyRunes, Runes: []rune("1")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("Z")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("6")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("4")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("4")},
		{Type: tea.KeyRunes, Runes: []rune("T")},
		{Type: tea.KeyRunes, Runes: []rune("1")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("3")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("Z")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("f")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("i")},
		{Type: tea.KeyRunes, Runes: []rune("l")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
		{Type: tea.KeyRunes, Runes: []rune("d")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("d")},
		{Type: tea.KeyRunes, Runes: []rune("i")},
		{Type: tea.KeyRunes, Runes: []rune("s")},
		{Type: tea.KeyRunes, Runes: []rune("p")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyRunes, Runes: []rune("c")},
		{Type: tea.KeyRunes, Runes: []rune("h")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRight},
		{Type: tea.KeyRunes, Runes: []rune("]")},
		{Type: tea.KeyRunes, Runes: []rune("]")},
		{Type: tea.KeyRunes, Runes: []rune(".")},
		{Type: tea.KeyRunes, Runes: []rune(".")},
		{Type: tea.KeyRunes, Runes: []rune("/")},
	} {
		updated, _ = page.Update(keyMsg)
		page = updated.(*DeadLettersPage)
	}

	updated, reloadCmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, reloadCmd)
	assert.Equal(t, "tx-2", page.appliedQuery)
	assert.Equal(t, "operator:alice", page.appliedManualReplayActor)
	assert.Equal(t, "2026-04-24T11:00:00Z", page.appliedDeadLetteredAfter)
	assert.Equal(t, "2026-04-24T12:30:00Z", page.appliedDeadLetteredBefore)
	assert.Equal(t, "failed", page.appliedDeadLetterReasonQuery)
	assert.Equal(t, "dispatch-2", page.appliedLatestDispatchReference)
	assert.Equal(t, deadLetterAdjudicationRelease, page.appliedAdjudication)
	assert.Equal(t, deadLetterSubtypeManualRetryRequested, page.appliedSubtype)
	assert.Equal(t, deadLetterFamilyManualRetry, page.appliedFamily)
	assert.Equal(t, deadLetterFamilyRetry, page.appliedAnyMatchFamily)
	assert.Equal(t, 1, listFn.called)

	updated, detailCmd = page.Update(reloadCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, DeadLetterListOptions{
		Query:                     "tx-2",
		Adjudication:              "release",
		LatestStatusSubtype:       "manual-retry-requested",
		LatestStatusSubtypeFamily: "manual-retry",
		AnyMatchFamily:            "retry",
		ManualReplayActor:         "operator:alice",
		DeadLetteredAfter:         "2026-04-24T11:00:00Z",
		DeadLetteredBefore:        "2026-04-24T12:30:00Z",
		DeadLetterReasonQuery:     "failed",
		LatestDispatchReference:   "dispatch-2",
	}, listFn.lastOptions)
	assert.Equal(t, 0, page.cursor)
	assert.Equal(t, "tx-2", page.selectedID)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-2", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
}

func TestDeadLettersPage_ApplyFiltersPreservesSelectionWhenPresent(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestManualReplayActor: "operator:bob", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "release", LatestManualReplayActor: "operator:alice", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-3", SubmissionReceiptID: "sub-3", Adjudication: "refund", LatestManualReplayActor: "operator:alice", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"}}, CanRetry: true},
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, CanRetry: true},
			"tx-3": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-3"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-3"}}, CanRetry: true},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.Equal(t, "tx-3", page.selectedID)

	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("p")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("l")},
		{Type: tea.KeyRunes, Runes: []rune("i")},
		{Type: tea.KeyRunes, Runes: []rune("c")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
	} {
		updated, _ = page.Update(keyMsg)
		page = updated.(*DeadLettersPage)
	}

	updated, reloadCmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, reloadCmd)

	updated, detailCmd = page.Update(reloadCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, "tx-3", page.selectedID)
	assert.Equal(t, 1, page.cursor)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-3", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
}

func TestDeadLettersPage_ApplyFiltersFallsBackToFirstRowWhenSelectionDisappears(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestManualReplayActor: "operator:bob", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "release", LatestManualReplayActor: "operator:alice", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-3", SubmissionReceiptID: "sub-3", Adjudication: "refund", LatestManualReplayActor: "operator:alice", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"}}, CanRetry: true},
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, CanRetry: true},
			"tx-3": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-3"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-3"}}, CanRetry: true},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.Equal(t, "tx-3", page.selectedID)

	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("p")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("b")},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("b")},
	} {
		updated, _ = page.Update(keyMsg)
		page = updated.(*DeadLettersPage)
	}

	updated, reloadCmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, reloadCmd)

	updated, detailCmd = page.Update(reloadCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, "tx-1", page.selectedID)
	assert.Equal(t, 0, page.cursor)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-1", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
}

func TestDeadLettersPage_EmptyFilteredResultClearsDetail(t *testing.T) {
	listFn := &mockDeadLetterListFn{}
	detailFn := &mockDeadLetterDetailFn{statusByID: map[string]postadjudicationstatus.TransactionStatus{}}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("n")},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("n")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
	} {
		updated, _ := page.Update(keyMsg)
		page = updated.(*DeadLettersPage)
	}
	updated, reloadCmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*DeadLettersPage)
	assert.Equal(t, "none", page.appliedQuery)
	require.NotNil(t, reloadCmd)

	updated, detailCmd := page.Update(reloadCmd())
	page = updated.(*DeadLettersPage)
	assert.Nil(t, detailCmd)
	assert.Empty(t, page.items)
	assert.Nil(t, page.detail)
	assert.Equal(t, DeadLetterListOptions{Query: "none"}, listFn.lastOptions)
	assert.Contains(t, page.View(), "No dead-letter backlog matches the current filters.")
}

func TestDeadLettersPage_NoBacklogDoesNotLoadDetail(t *testing.T) {
	listFn := &mockDeadLetterListFn{}
	detailFn := &mockDeadLetterDetailFn{statusByID: map[string]postadjudicationstatus.TransactionStatus{}}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)

	assert.Empty(t, page.items)
	assert.Nil(t, detailCmd)
	assert.Nil(t, page.detail)
	assert.Empty(t, detailFn.calls)
	assert.Contains(t, page.View(), "No current dead-letter backlog.")
}

func TestDeadLettersPage_ShowsLoadErrors(t *testing.T) {
	page := NewDeadLettersPage((&mockDeadLetterListFn{err: errors.New("boom")}).call, nil)
	updated, _ := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	assert.Contains(t, page.View(), "Failed to load dead letters")
}

func TestDeadLettersPage_ShowsDetailErrors(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 2, LatestDeadLetterReason: "failed", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{},
		errByID:    map[string]error{"tx-1": errors.New("detail exploded")},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	assert.Contains(t, page.View(), "Failed to load detail")
}

func TestDeadLettersPage_ViewIncludesBackgroundTaskWhenPresent(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 3, LatestDeadLetterReason: "terminal failure", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestDeadLetterReason:  "terminal failure",
					LatestRetryAttempt:      3,
					LatestDispatchReference: "dispatch-9",
				},
				LatestBackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{
					TaskID:       "task-9",
					Status:       "retrying",
					AttemptCount: 2,
					NextRetryAt:  "2026-04-24T12:30:00Z",
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "release",
			},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	view := page.View()
	assert.Contains(t, view, "Background task: task-9")
	assert.Contains(t, view, "Task status: retrying")
	assert.Contains(t, view, "Query: all")
	assert.Contains(t, view, "Manual replay actor: all")
	assert.Contains(t, view, "Dead-lettered after: all")
	assert.Contains(t, view, "Dead-lettered before: all")
	assert.Contains(t, view, "Dead-letter reason: all")
	assert.Contains(t, view, "Dispatch reference: all")
	assert.Contains(t, view, "Adjudication: all")
	assert.Contains(t, view, "Latest subtype: all")
	assert.Contains(t, view, "Latest family: all")
	assert.Contains(t, view, "Any-match family: all")
	assert.Contains(t, view, "Retry action: ready (press r to request retry)")
}

func TestDeadLettersPage_RetrySelectedShowsSuccessMessage(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 3, LatestDeadLetterReason: "terminal failure", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "release",
			},
		},
	}
	retryFn := &mockDeadLetterRetryFn{}

	page := NewDeadLettersPage(listFn.call, detailFn.call, retryFn.call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, retryCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.Nil(t, retryCmd)
	assert.Empty(t, retryFn.calls)
	assert.Contains(t, page.View(), "Retry action: confirm request (press r again)")

	updated, retryCmd = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, retryCmd)
	assert.Contains(t, page.View(), "Retry action: requesting retry...")

	updated, duplicateCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.Nil(t, duplicateCmd)
	assert.Empty(t, retryFn.calls)
	assert.Contains(t, page.View(), "Retry action: requesting retry...")

	updated, refreshCmd := page.Update(retryCmd())
	page = updated.(*DeadLettersPage)
	assert.Equal(t, []string{"tx-1"}, retryFn.calls)
	require.NotNil(t, refreshCmd)
	assert.Contains(t, page.View(), "Retry request accepted for tx-1. Refreshing backlog and detail.")
	assert.Empty(t, page.retryConfirmID)

	updated, detailCmd = page.Update(refreshCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-1", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
	assert.Equal(t, 2, listFn.called)
	assert.Equal(t, []string{"tx-1", "tx-1"}, detailFn.calls)
	assert.Contains(t, page.View(), "Retry action: ready (press r to request retry)")
	assert.Contains(t, page.View(), "Retry request accepted for tx-1. Refreshing backlog and detail.")
}

func TestDeadLettersPage_ViewIncludesSummaryStrip(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{
				TransactionReceiptID:      "tx-1",
				SubmissionReceiptID:       "sub-1",
				Adjudication:              "release",
				LatestDeadLetterReason:    "worker exhausted",
				LatestStatusSubtypeFamily: "retry",
				LatestManualReplayActor:   "operator:alice",
				LatestDispatchReference:   "dispatch-a",
				IsDeadLettered:            true,
				CanRetry:                  true,
			},
			{
				TransactionReceiptID:      "tx-2",
				SubmissionReceiptID:       "sub-2",
				Adjudication:              "refund",
				LatestDeadLetterReason:    "invalid receipt",
				LatestStatusSubtypeFamily: "manual-retry",
				LatestManualReplayActor:   "operator:bob",
				LatestDispatchReference:   "dispatch-b",
				IsDeadLettered:            true,
				CanRetry:                  false,
			},
			{
				TransactionReceiptID:      "tx-3",
				SubmissionReceiptID:       "sub-3",
				Adjudication:              "release",
				LatestDeadLetterReason:    "worker exhausted",
				LatestStatusSubtypeFamily: "dead-letter",
				LatestManualReplayActor:   "operator:alice",
				LatestDispatchReference:   "dispatch-a",
				IsDeadLettered:            true,
				CanRetry:                  true,
			},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"}}, CanRetry: true},
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, CanRetry: false},
			"tx-3": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-3"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-3"}}, CanRetry: true},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	view := page.View()
	assert.Contains(t, view, "dead letters: 3")
	assert.Contains(t, view, "retryable: 2")
	assert.Contains(t, view, "release/refund: 2/1")
	assert.Contains(t, view, "retry/manual/dead: 1/1/1")
	assert.Contains(t, view, "reasons: worker exhausted(2), invalid receipt(1)")
	assert.Contains(t, view, "actors: operator:alice(2), operator:bob(1)")
	assert.Contains(t, view, "dispatch: dispatch-a(2), dispatch-b(1)")
}

func TestSummarizeDeadLetters_TopReasonsActorsAndDispatches(t *testing.T) {
	summary := summarizeDeadLetters([]postadjudicationstatus.DeadLetterBacklogEntry{
		{LatestDeadLetterReason: "worker exhausted", LatestManualReplayActor: "operator:bob", LatestDispatchReference: "dispatch-c"},
		{LatestDeadLetterReason: "worker exhausted", LatestManualReplayActor: "operator:bob", LatestDispatchReference: "dispatch-c"},
		{LatestDeadLetterReason: "invalid receipt", LatestManualReplayActor: "operator:alice", LatestDispatchReference: "dispatch-a"},
		{LatestDeadLetterReason: "timeout", LatestManualReplayActor: "operator:zoe", LatestDispatchReference: "dispatch-e"},
		{LatestDeadLetterReason: "bad signature", LatestManualReplayActor: "operator:carl", LatestDispatchReference: "dispatch-b"},
		{LatestDeadLetterReason: "queue saturated", LatestManualReplayActor: "operator:alice", LatestDispatchReference: "dispatch-a"},
		{LatestDeadLetterReason: "queue saturated", LatestManualReplayActor: "operator:alice", LatestDispatchReference: "dispatch-a"},
		{LatestDeadLetterReason: "policy denied", LatestManualReplayActor: "operator:dana", LatestDispatchReference: "dispatch-d"},
		{LatestDeadLetterReason: ""},
	})

	require.Len(t, summary.topReasons, 5)
	assert.Equal(t, []deadLetterReasonSummaryItem{
		{reason: "queue saturated", count: 2},
		{reason: "worker exhausted", count: 2},
		{reason: "bad signature", count: 1},
		{reason: "invalid receipt", count: 1},
		{reason: "policy denied", count: 1},
	}, summary.topReasons)

	require.Len(t, summary.topActors, 5)
	assert.Equal(t, []deadLetterActorSummaryItem{
		{actor: "operator:alice", count: 3},
		{actor: "operator:bob", count: 2},
		{actor: "operator:carl", count: 1},
		{actor: "operator:dana", count: 1},
		{actor: "operator:zoe", count: 1},
	}, summary.topActors)

	require.Len(t, summary.topDispatches, 5)
	assert.Equal(t, []deadLetterDispatchSummaryItem{
		{dispatchReference: "dispatch-a", count: 3},
		{dispatchReference: "dispatch-c", count: 2},
		{dispatchReference: "dispatch-b", count: 1},
		{dispatchReference: "dispatch-d", count: 1},
		{dispatchReference: "dispatch-e", count: 1},
	}, summary.topDispatches)
}

func TestDeadLettersPage_SummaryStripRecomputesAcrossReloadPaths(t *testing.T) {
	initialItems := []postadjudicationstatus.DeadLetterBacklogEntry{
		{
			TransactionReceiptID:      "tx-1",
			SubmissionReceiptID:       "sub-1",
			Adjudication:              "release",
			LatestDeadLetterReason:    "worker exhausted",
			LatestStatusSubtypeFamily: "retry",
			LatestManualReplayActor:   "operator:alice",
			LatestDispatchReference:   "dispatch-a",
			IsDeadLettered:            true,
			CanRetry:                  true,
		},
		{
			TransactionReceiptID:      "tx-2",
			SubmissionReceiptID:       "sub-2",
			Adjudication:              "refund",
			LatestDeadLetterReason:    "invalid receipt",
			LatestStatusSubtypeFamily: "dead-letter",
			LatestManualReplayActor:   "operator:bob",
			LatestDispatchReference:   "dispatch-b",
			IsDeadLettered:            true,
			CanRetry:                  false,
		},
	}
	filteredItems := []postadjudicationstatus.DeadLetterBacklogEntry{
		{
			TransactionReceiptID:      "tx-1",
			SubmissionReceiptID:       "sub-1",
			Adjudication:              "release",
			LatestDeadLetterReason:    "manual gate",
			LatestStatusSubtypeFamily: "manual-retry",
			LatestManualReplayActor:   "operator:alice",
			LatestDispatchReference:   "dispatch-c",
			IsDeadLettered:            true,
			CanRetry:                  true,
		},
	}
	resetItems := []postadjudicationstatus.DeadLetterBacklogEntry{
		{
			TransactionReceiptID:      "tx-1",
			SubmissionReceiptID:       "sub-1",
			Adjudication:              "release",
			LatestDeadLetterReason:    "worker exhausted",
			LatestStatusSubtypeFamily: "retry",
			LatestManualReplayActor:   "operator:alice",
			LatestDispatchReference:   "dispatch-a",
			IsDeadLettered:            true,
			CanRetry:                  true,
		},
		{
			TransactionReceiptID:      "tx-2",
			SubmissionReceiptID:       "sub-2",
			Adjudication:              "refund",
			LatestDeadLetterReason:    "manual gate",
			LatestStatusSubtypeFamily: "manual-retry",
			LatestManualReplayActor:   "operator:bob",
			LatestDispatchReference:   "dispatch-c",
			IsDeadLettered:            true,
			CanRetry:                  true,
		},
		{
			TransactionReceiptID:      "tx-3",
			SubmissionReceiptID:       "sub-3",
			Adjudication:              "release",
			LatestDeadLetterReason:    "invalid receipt",
			LatestStatusSubtypeFamily: "dead-letter",
			LatestManualReplayActor:   "operator:carol",
			LatestDispatchReference:   "dispatch-b",
			IsDeadLettered:            true,
			CanRetry:                  false,
		},
	}
	postRetryItems := []postadjudicationstatus.DeadLetterBacklogEntry{
		{
			TransactionReceiptID:      "tx-2",
			SubmissionReceiptID:       "sub-2",
			Adjudication:              "refund",
			LatestDeadLetterReason:    "manual gate",
			LatestStatusSubtypeFamily: "manual-retry",
			LatestManualReplayActor:   "operator:bob",
			LatestDispatchReference:   "dispatch-c",
			IsDeadLettered:            true,
			CanRetry:                  true,
		},
		{
			TransactionReceiptID:      "tx-3",
			SubmissionReceiptID:       "sub-3",
			Adjudication:              "release",
			LatestDeadLetterReason:    "invalid receipt",
			LatestStatusSubtypeFamily: "dead-letter",
			LatestManualReplayActor:   "operator:carol",
			LatestDispatchReference:   "dispatch-b",
			IsDeadLettered:            true,
			CanRetry:                  false,
		},
	}

	listFn := &sequenceDeadLetterListFn{
		results: [][]postadjudicationstatus.DeadLetterBacklogEntry{
			initialItems,
			filteredItems,
			resetItems,
			postRetryItems,
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"}}, CanRetry: true},
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, CanRetry: true},
			"tx-3": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-3"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-3"}}, CanRetry: false},
		},
	}
	retryFn := &mockDeadLetterRetryFn{}

	page := NewDeadLettersPage(listFn.call, detailFn.call, retryFn.call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	assert.Contains(t, page.View(), "dead letters: 2")
	assert.Contains(t, page.View(), "retryable: 1")
	assert.Contains(t, page.View(), "release/refund: 1/1")
	assert.Contains(t, page.View(), "retry/manual/dead: 1/0/1")
	assert.Contains(t, page.View(), "reasons: invalid receipt(1), worker exhausted(1)")
	assert.Contains(t, page.View(), "actors: operator:alice(1), operator:bob(1)")
	assert.Contains(t, page.View(), "dispatch: dispatch-a(1), dispatch-b(1)")

	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("p")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("l")},
		{Type: tea.KeyRunes, Runes: []rune("i")},
		{Type: tea.KeyRunes, Runes: []rune("c")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
	} {
		updated, _ = page.Update(keyMsg)
		page = updated.(*DeadLettersPage)
	}
	updated, reloadCmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, reloadCmd)
	updated, detailCmd = page.Update(reloadCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	assert.Contains(t, page.View(), "dead letters: 1")
	assert.Contains(t, page.View(), "retryable: 1")
	assert.Contains(t, page.View(), "release/refund: 1/0")
	assert.Contains(t, page.View(), "retry/manual/dead: 0/1/0")
	assert.Contains(t, page.View(), "reasons: manual gate(1)")
	assert.Contains(t, page.View(), "actors: operator:alice(1)")
	assert.Contains(t, page.View(), "dispatch: dispatch-c(1)")

	updated, reloadCmd = page.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, reloadCmd)
	updated, detailCmd = page.Update(reloadCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	assert.Contains(t, page.View(), "dead letters: 3")
	assert.Contains(t, page.View(), "retryable: 2")
	assert.Contains(t, page.View(), "release/refund: 2/1")
	assert.Contains(t, page.View(), "retry/manual/dead: 1/1/1")
	assert.Contains(t, page.View(), "reasons: invalid receipt(1), manual gate(1), worker exhausted(1)")
	assert.Contains(t, page.View(), "actors: operator:alice(1), operator:bob(1), operator:carol(1)")
	assert.Contains(t, page.View(), "dispatch: dispatch-a(1), dispatch-b(1), dispatch-c(1)")

	updated, retryCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	require.Nil(t, retryCmd)
	updated, retryCmd = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, retryCmd)
	updated, refreshCmd := page.Update(retryCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, refreshCmd)
	updated, detailCmd = page.Update(refreshCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	assert.Contains(t, page.View(), "dead letters: 2")
	assert.Contains(t, page.View(), "retryable: 1")
	assert.Contains(t, page.View(), "release/refund: 1/1")
	assert.Contains(t, page.View(), "retry/manual/dead: 0/1/1")
	assert.Contains(t, page.View(), "reasons: invalid receipt(1), manual gate(1)")
	assert.Contains(t, page.View(), "actors: operator:bob(1), operator:carol(1)")
	assert.Contains(t, page.View(), "dispatch: dispatch-b(1), dispatch-c(1)")
}

func TestDeadLettersPage_RetrySelectedShowsFailureMessage(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "refund", LatestRetryAttempt: 2, LatestDeadLetterReason: "refund failed", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "refund",
			},
		},
	}
	retryFn := &mockDeadLetterRetryFn{errByID: map[string]error{"tx-1": errors.New("policy denied")}}

	page := NewDeadLettersPage(listFn.call, detailFn.call, retryFn.call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, retryCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.Nil(t, retryCmd)

	updated, retryCmd = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, retryCmd)
	assert.Contains(t, page.View(), "Retry action: requesting retry...")

	updated, duplicateCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.Nil(t, duplicateCmd)
	assert.Empty(t, retryFn.calls)

	updated, refreshCmd := page.Update(retryCmd())
	page = updated.(*DeadLettersPage)
	assert.Nil(t, refreshCmd)
	assert.Equal(t, []string{"tx-1"}, retryFn.calls)
	assert.Contains(t, page.View(), "Retry request failed for tx-1: policy denied")
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-1", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
	assert.Equal(t, 1, listFn.called)
	assert.Contains(t, page.View(), "Retry action: ready (press r to request retry)")
}

func TestDeadLettersPage_RetryIgnoredWhenDetailIsNotRetryable(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 1, LatestDeadLetterReason: "inactive", IsDeadLettered: true, CanRetry: false},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				IsDeadLettered: true,
				CanRetry:       false,
				Adjudication:   "release",
			},
		},
	}
	retryFn := &mockDeadLetterRetryFn{}

	page := NewDeadLettersPage(listFn.call, detailFn.call, retryFn.call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, retryCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.Nil(t, retryCmd)
	assert.Empty(t, retryFn.calls)
	assert.Contains(t, page.View(), "Retry action: disabled")
}

func TestDeadLettersPage_RetryConfirmClearsOnEscape(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 3, LatestDeadLetterReason: "terminal failure", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "release",
			},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.True(t, page.retryConfirmActive())

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyEsc})
	page = updated.(*DeadLettersPage)
	assert.False(t, page.retryConfirmActive())
	assert.Contains(t, page.View(), "Retry action: ready (press r to request retry)")
}

func TestDeadLettersPage_RetryConfirmClearsOnSelectionChange(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 3, LatestDeadLetterReason: "terminal failure", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "refund", LatestRetryAttempt: 1, LatestDeadLetterReason: "other failure", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "release",
			},
			"tx-2": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"},
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "refund",
			},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.True(t, page.retryConfirmActive())

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.False(t, page.retryConfirmActive())
	assert.Equal(t, "tx-2", page.selectedID)
}

func TestDeadLettersPage_RetryConfirmClearsOnFilterApply(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 3, LatestDeadLetterReason: "terminal failure", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "release",
			},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.True(t, page.retryConfirmActive())

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	page = updated.(*DeadLettersPage)
	assert.False(t, page.retryConfirmActive())

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.True(t, page.retryConfirmActive())

	updated, reloadCmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, reloadCmd)
	assert.False(t, page.retryConfirmActive())
}

func TestDeadLettersPage_ResetShortcutClearsDraftAppliedAndReloads(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "release", LatestStatusSubtype: "manual-retry-requested", LatestStatusSubtypeFamily: "manual-retry", AnyMatchFamilies: []string{"manual-retry", "retry"}, LatestDispatchReference: "dispatch-2", LatestManualReplayActor: "operator:alice", LatestDeadLetteredAt: "2026-04-24T12:00:00Z", LatestRetryAttempt: 4, LatestDeadLetterReason: "release failed", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-3", SubmissionReceiptID: "sub-3", Adjudication: "refund", LatestStatusSubtype: "dead-lettered", LatestStatusSubtypeFamily: "dead-letter", AnyMatchFamilies: []string{"dead-letter"}, LatestDispatchReference: "dispatch-3", LatestManualReplayActor: "operator:bob", LatestDeadLetteredAt: "2026-04-24T10:00:00Z", LatestRetryAttempt: 2, LatestDeadLetterReason: "refund failed", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, CanRetry: true, Adjudication: "release", IsDeadLettered: true},
			"tx-3": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-3"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-3"}}, CanRetry: true, Adjudication: "refund", IsDeadLettered: true},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyRunes, Runes: []rune("x")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("p")},
		{Type: tea.KeyRunes, Runes: []rune("1")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("6")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("4")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("4")},
		{Type: tea.KeyRunes, Runes: []rune("T")},
		{Type: tea.KeyRunes, Runes: []rune("1")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("Z")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("6")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("4")},
		{Type: tea.KeyRunes, Runes: []rune("-")},
		{Type: tea.KeyRunes, Runes: []rune("2")},
		{Type: tea.KeyRunes, Runes: []rune("4")},
		{Type: tea.KeyRunes, Runes: []rune("T")},
		{Type: tea.KeyRunes, Runes: []rune("1")},
		{Type: tea.KeyRunes, Runes: []rune("1")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("0")},
		{Type: tea.KeyRunes, Runes: []rune("Z")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("f")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("i")},
		{Type: tea.KeyRunes, Runes: []rune("l")},
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("d")},
		{Type: tea.KeyRunes, Runes: []rune("i")},
		{Type: tea.KeyRunes, Runes: []rune("s")},
		{Type: tea.KeyRight},
		{Type: tea.KeyRunes, Runes: []rune("]")},
		{Type: tea.KeyRunes, Runes: []rune(".")},
		{Type: tea.KeyRunes, Runes: []rune("/")},
	} {
		updated, _ = page.Update(keyMsg)
		page = updated.(*DeadLettersPage)
	}

	updated, reloadCmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, reloadCmd)
	updated, detailCmd = page.Update(reloadCmd())
	page = updated.(*DeadLettersPage)
	if detailCmd != nil {
		updated, _ = page.Update(detailCmd())
		page = updated.(*DeadLettersPage)
	}

	assert.NotEmpty(t, page.appliedQuery)
	assert.NotEmpty(t, page.appliedManualReplayActor)
	assert.NotEmpty(t, page.appliedDeadLetteredAfter)
	assert.NotEmpty(t, page.appliedDeadLetteredBefore)
	assert.NotEmpty(t, page.appliedDeadLetterReasonQuery)
	assert.NotEmpty(t, page.appliedLatestDispatchReference)
	assert.NotEqual(t, deadLetterAdjudicationAll, page.appliedAdjudication)
	assert.NotEqual(t, deadLetterSubtypeAll, page.appliedSubtype)
	assert.NotEqual(t, deadLetterFamilyAll, page.appliedFamily)
	assert.NotEqual(t, deadLetterFamilyAll, page.appliedAnyMatchFamily)

	updated, resetCmd := page.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, resetCmd)
	assert.Equal(t, deadLetterTextFieldQuery, page.activeTextField)
	assert.Empty(t, page.queryDraft)
	assert.Empty(t, page.manualReplayActorDraft)
	assert.Empty(t, page.deadLetteredAfterDraft)
	assert.Empty(t, page.deadLetteredBeforeDraft)
	assert.Empty(t, page.deadLetterReasonQueryDraft)
	assert.Empty(t, page.latestDispatchReferenceDraft)
	assert.Empty(t, page.appliedQuery)
	assert.Empty(t, page.appliedManualReplayActor)
	assert.Empty(t, page.appliedDeadLetteredAfter)
	assert.Empty(t, page.appliedDeadLetteredBefore)
	assert.Empty(t, page.appliedDeadLetterReasonQuery)
	assert.Empty(t, page.appliedLatestDispatchReference)
	assert.Equal(t, deadLetterAdjudicationAll, page.appliedAdjudication)
	assert.Equal(t, deadLetterSubtypeAll, page.appliedSubtype)
	assert.Equal(t, deadLetterFamilyAll, page.appliedFamily)
	assert.Equal(t, deadLetterFamilyAll, page.appliedAnyMatchFamily)

	updated, detailCmd = page.Update(resetCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, DeadLetterListOptions{}, listFn.lastOptions)
	assert.Equal(t, 0, page.cursor)
	assert.Equal(t, "tx-2", page.selectedID)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-2", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
}

func TestDeadLettersPage_ResetShortcutClearsRetryConfirmState(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 3, LatestDeadLetterReason: "terminal failure", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "release",
			},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	require.True(t, page.retryConfirmActive())

	updated, resetCmd := page.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, resetCmd)
	assert.False(t, page.retryConfirmActive())
}

func TestDeadLettersPage_ResetShortcutIgnoredWhileRetryRunning(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestRetryAttempt: 3, LatestDeadLetterReason: "terminal failure", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
					SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
				},
				IsDeadLettered: true,
				CanRetry:       true,
				Adjudication:   "release",
			},
		},
	}
	retryFn := &mockDeadLetterRetryFn{}

	page := NewDeadLettersPage(listFn.call, detailFn.call, retryFn.call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	updated, retryCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, retryCmd)
	require.True(t, page.retryRunning())

	page.queryDraft = "stale"
	page.appliedQuery = "stale"
	page.retryConfirmID = "tx-1"

	updated, resetCmd := page.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	page = updated.(*DeadLettersPage)
	assert.Nil(t, resetCmd)
	assert.Equal(t, "stale", page.queryDraft)
	assert.Equal(t, "stale", page.appliedQuery)
	assert.Equal(t, "tx-1", page.retryConfirmID)
	assert.True(t, page.retryRunning())
	assert.Equal(t, 1, listFn.called)
}

func TestDeadLettersPage_ResetShortcutPreservesSelectionWhenPresent(t *testing.T) {
	listFn := &mockDeadLetterListFn{
		items: []postadjudicationstatus.DeadLetterBacklogEntry{
			{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", LatestManualReplayActor: "operator:bob", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "release", LatestManualReplayActor: "operator:alice", IsDeadLettered: true, CanRetry: true},
			{TransactionReceiptID: "tx-3", SubmissionReceiptID: "sub-3", Adjudication: "refund", LatestManualReplayActor: "operator:alice", IsDeadLettered: true, CanRetry: true},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"}}, CanRetry: true},
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, CanRetry: true},
			"tx-3": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-3"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-3"}}, CanRetry: true},
		},
	}

	page := NewDeadLettersPage(listFn.call, detailFn.call, (&mockDeadLetterRetryFn{}).call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.Equal(t, "tx-2", page.selectedID)

	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyTab},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("p")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyRunes, Runes: []rune("o")},
		{Type: tea.KeyRunes, Runes: []rune("r")},
		{Type: tea.KeyRunes, Runes: []rune(":")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
		{Type: tea.KeyRunes, Runes: []rune("l")},
		{Type: tea.KeyRunes, Runes: []rune("i")},
		{Type: tea.KeyRunes, Runes: []rune("c")},
		{Type: tea.KeyRunes, Runes: []rune("e")},
	} {
		updated, _ = page.Update(keyMsg)
		page = updated.(*DeadLettersPage)
	}

	updated, reloadCmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*DeadLettersPage)
	updated, detailCmd = page.Update(reloadCmd())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.Equal(t, "tx-2", page.selectedID)

	updated, resetCmd := page.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, resetCmd)

	updated, detailCmd = page.Update(resetCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, "tx-2", page.selectedID)
	assert.Equal(t, 1, page.cursor)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-2", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
}

func TestDeadLettersPage_RetrySuccessRefreshPreservesCurrentSelectionWhenPresent(t *testing.T) {
	listFn := &sequenceDeadLetterListFn{
		results: [][]postadjudicationstatus.DeadLetterBacklogEntry{
			{
				{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", IsDeadLettered: true, CanRetry: true},
				{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "refund", IsDeadLettered: true, CanRetry: true},
			},
			{
				{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", IsDeadLettered: true, CanRetry: true},
				{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "refund", IsDeadLettered: true, CanRetry: true},
			},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"}}, CanRetry: true},
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, CanRetry: true},
		},
	}
	retryFn := &mockDeadLetterRetryFn{}

	page := NewDeadLettersPage(listFn.call, detailFn.call, retryFn.call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, retryCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.Nil(t, retryCmd)

	updated, retryCmd = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, retryCmd)
	require.Equal(t, "tx-1", page.retryRunningID)

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, "tx-2", page.selectedID)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, refreshCmd := page.Update(retryCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, refreshCmd)
	assert.Equal(t, []string{"tx-1"}, retryFn.calls)

	updated, detailCmd = page.Update(refreshCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, "tx-2", page.selectedID)
	assert.Equal(t, 1, page.cursor)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-2", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
}

func TestDeadLettersPage_RetrySuccessRefreshFallsBackToFirstRowWhenSelectionDisappears(t *testing.T) {
	listFn := &sequenceDeadLetterListFn{
		results: [][]postadjudicationstatus.DeadLetterBacklogEntry{
			{
				{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", IsDeadLettered: true, CanRetry: true},
				{TransactionReceiptID: "tx-2", SubmissionReceiptID: "sub-2", Adjudication: "refund", IsDeadLettered: true, CanRetry: true},
			},
			{
				{TransactionReceiptID: "tx-1", SubmissionReceiptID: "sub-1", Adjudication: "release", IsDeadLettered: true, CanRetry: true},
			},
		},
	}
	detailFn := &mockDeadLetterDetailFn{
		statusByID: map[string]postadjudicationstatus.TransactionStatus{
			"tx-1": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"}}, CanRetry: true},
			"tx-2": {CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-2"}, SubmissionReceipt: receipts.SubmissionReceipt{SubmissionReceiptID: "sub-2"}}, CanRetry: true},
		},
	}
	retryFn := &mockDeadLetterRetryFn{}

	page := NewDeadLettersPage(listFn.call, detailFn.call, retryFn.call)
	updated, detailCmd := page.Update(page.Activate()())
	page = updated.(*DeadLettersPage)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, retryCmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	assert.Nil(t, retryCmd)

	updated, retryCmd = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, retryCmd)

	updated, detailCmd = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, "tx-2", page.selectedID)
	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)

	updated, refreshCmd := page.Update(retryCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, refreshCmd)

	updated, detailCmd = page.Update(refreshCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, detailCmd)
	assert.Equal(t, "tx-1", page.selectedID)
	assert.Equal(t, 0, page.cursor)

	updated, _ = page.Update(detailCmd())
	page = updated.(*DeadLettersPage)
	require.NotNil(t, page.detail)
	assert.Equal(t, "tx-1", page.detail.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
}
