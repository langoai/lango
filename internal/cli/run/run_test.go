package run

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/runledger"
)

type stubRunLedgerStore struct {
	runs   []runledger.RunSummary
	events []runledger.JournalEvent
}

func (s stubRunLedgerStore) AppendJournalEvent(context.Context, runledger.JournalEvent) error {
	return errors.New("not implemented")
}
func (s stubRunLedgerStore) GetJournalEvents(context.Context, string) ([]runledger.JournalEvent, error) {
	return s.events, nil
}
func (s stubRunLedgerStore) GetJournalEventsSince(context.Context, string, int64) ([]runledger.JournalEvent, error) {
	return nil, errors.New("not implemented")
}
func (s stubRunLedgerStore) MaterializeRunSnapshot(context.Context, string) (*runledger.RunSnapshot, error) {
	return nil, errors.New("not implemented")
}
func (s stubRunLedgerStore) RecordValidationResult(context.Context, string, string, runledger.ValidationResult) error {
	return errors.New("not implemented")
}
func (s stubRunLedgerStore) GetCachedSnapshot(context.Context, string) (*runledger.RunSnapshot, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (s stubRunLedgerStore) UpdateCachedSnapshot(context.Context, *runledger.RunSnapshot) error {
	return errors.New("not implemented")
}
func (s stubRunLedgerStore) ListRuns(context.Context, int) ([]runledger.RunSummary, error) {
	return s.runs, nil
}
func (s stubRunLedgerStore) GetRunSnapshot(context.Context, string) (*runledger.RunSnapshot, error) {
	return nil, errors.New("not implemented")
}
func (s stubRunLedgerStore) ListRunSummariesBySession(context.Context, string, int) ([]runledger.RunSummary, error) {
	return nil, errors.New("not implemented")
}
func (s stubRunLedgerStore) MaxJournalSeqForSession(context.Context, string) (int64, error) {
	return 0, errors.New("not implemented")
}
func (s stubRunLedgerStore) PruneOldRuns(context.Context, int) error {
	return errors.New("not implemented")
}

func executeRunCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestRunList_WritesToCommandOutput(t *testing.T) {
	orig := runLedgerStoreFromBoot
	t.Cleanup(func() { runLedgerStoreFromBoot = orig })
	runLedgerStoreFromBoot = func(boot *bootstrap.Result) runledger.RunLedgerStore {
		return stubRunLedgerStore{runs: []runledger.RunSummary{{
			RunID:          "run-1",
			Status:         runledger.RunStatusRunning,
			Goal:           "Ship feature",
			CompletedSteps: 1,
			TotalSteps:     3,
		}}}
	}

	cmd := NewRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.RunLedger.Enabled = true
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeRunCommand(t, cmd, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "run-1")
	assert.Contains(t, out, "Ship feature")
}

func TestRunList_JSONAndInvalidOutput(t *testing.T) {
	orig := runLedgerStoreFromBoot
	t.Cleanup(func() { runLedgerStoreFromBoot = orig })
	runLedgerStoreFromBoot = func(boot *bootstrap.Result) runledger.RunLedgerStore {
		return stubRunLedgerStore{runs: []runledger.RunSummary{{
			RunID:          "run-1",
			Status:         runledger.RunStatusRunning,
			Goal:           "Ship feature",
			CompletedSteps: 1,
			TotalSteps:     3,
		}}}
	}

	cmd := NewRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.RunLedger.Enabled = true
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeRunCommand(t, cmd, "list", "--output", "json")
	require.NoError(t, err)
	var payload []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Len(t, payload, 1)
	assert.Equal(t, "run-1", payload[0]["run_id"])

	_, err = executeRunCommand(t, cmd, "list", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestRunStatus_WritesToCommandOutput(t *testing.T) {
	cmd := NewRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.RunLedger.Enabled = true
		cfg.RunLedger.PlannerMaxRetries = 7
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeRunCommand(t, cmd, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "RunLedger Configuration")
	assert.Contains(t, out, "Planner Retries:    7")
}

func TestRunStatus_JSONAndInvalidOutput(t *testing.T) {
	cmd := NewRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.RunLedger.Enabled = true
		cfg.RunLedger.PlannerMaxRetries = 7
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeRunCommand(t, cmd, "status", "--output", "json")
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, true, payload["enabled"])
	assert.Equal(t, float64(7), payload["plannerMaxRetries"])

	_, err = executeRunCommand(t, cmd, "status", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestRunJournal_WritesToCommandOutput(t *testing.T) {
	orig := runLedgerStoreFromBoot
	t.Cleanup(func() { runLedgerStoreFromBoot = orig })
	runLedgerStoreFromBoot = func(boot *bootstrap.Result) runledger.RunLedgerStore {
		return stubRunLedgerStore{events: []runledger.JournalEvent{{
			RunID:     "run-1",
			Seq:       1,
			Type:      runledger.EventRunCreated,
			Timestamp: time.Now(),
			Payload:   []byte(`{"goal":"Ship feature"}`),
		}}}
	}

	cmd := NewRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.RunLedger.Enabled = true
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeRunCommand(t, cmd, "journal", "run-1")
	require.NoError(t, err)
	assert.Contains(t, out, "1\trun_created")
	assert.Contains(t, out, `"goal":"Ship feature"`)
}

func TestRunJournal_JSONLimitAndInvalidOutput(t *testing.T) {
	orig := runLedgerStoreFromBoot
	t.Cleanup(func() { runLedgerStoreFromBoot = orig })
	runLedgerStoreFromBoot = func(boot *bootstrap.Result) runledger.RunLedgerStore {
		return stubRunLedgerStore{events: []runledger.JournalEvent{
			{
				RunID:     "run-1",
				Seq:       1,
				Type:      runledger.EventRunCreated,
				Timestamp: time.Now(),
				Payload:   []byte(`{"goal":"Ship feature"}`),
			},
			{
				RunID:     "run-1",
				Seq:       2,
				Type:      runledger.EventStepStarted,
				Timestamp: time.Now(),
				Payload:   []byte(`{"step":"collect"}`),
			},
		}}
	}

	cmd := NewRunCmd(func() (*bootstrap.Result, error) {
		cfg := config.DefaultConfig()
		cfg.RunLedger.Enabled = true
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeRunCommand(t, cmd, "journal", "run-1", "--limit", "1", "--output", "json")
	require.NoError(t, err)
	var payload []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Len(t, payload, 1)
	assert.Equal(t, float64(1), payload[0]["seq"])

	_, err = executeRunCommand(t, cmd, "journal", "run-1", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}
