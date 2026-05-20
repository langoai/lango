package provenance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/enttest"
	provenancepkg "github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/storage"
)

func TestCheckpointListRequiresFilterAndPrintsEmptyResults(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no filter",
			args: []string{"checkpoint", "list"},
			want: "Specify --run or --session to filter checkpoints.",
		},
		{
			name: "empty run",
			args: []string{"checkpoint", "list", "--run", "missing-run"},
			want: "No checkpoints found.",
		},
		{
			name: "empty session",
			args: []string{"checkpoint", "list", "--session", "missing-session"},
			want: "No checkpoints found.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewProvenanceCmd(enabledEntBootLoader(t, newProvenanceTestClient(t)))
			out, err := executeProvenanceCommand(cmd, tt.args...)

			require.NoError(t, err)
			assert.Equal(t, tt.want, strings.TrimSpace(out))
		})
	}
}

func TestCheckpointListPrintsRunAndSessionRows(t *testing.T) {
	created := time.Date(2026, 5, 20, 9, 30, 15, 0, time.UTC)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "run filter",
			args: []string{"checkpoint", "list", "--run", "run-a"},
			want: "11111111\tmanual\tfirst\tseq=7\t2026-05-20 09:30:15",
		},
		{
			name: "session filter",
			args: []string{"checkpoint", "list", "--session", "session-a"},
			want: strings.Join([]string{
				"22222222\tpolicy_applied\tsecond\tseq=9\t2026-05-20 09:31:15",
				"11111111\tmanual\tfirst\tseq=7\t2026-05-20 09:30:15",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newProvenanceTestClient(t)
			saveCheckpointFixtures(t, client, []provenancepkg.Checkpoint{
				{
					ID:         "11111111-1111-1111-1111-111111111111",
					SessionKey: "session-a",
					RunID:      "run-a",
					Label:      "first",
					Trigger:    provenancepkg.TriggerManual,
					JournalSeq: 7,
					CreatedAt:  created,
				},
				{
					ID:         "22222222-2222-2222-2222-222222222222",
					SessionKey: "session-a",
					RunID:      "run-b",
					Label:      "second",
					Trigger:    provenancepkg.TriggerPolicy,
					JournalSeq: 9,
					CreatedAt:  created.Add(time.Minute),
				},
			})
			cmd := NewProvenanceCmd(enabledEntBootLoader(t, client))
			out, err := executeProvenanceCommand(cmd, tt.args...)

			require.NoError(t, err)
			assert.Equal(t, tt.want, strings.TrimSpace(out))
		})
	}
}

func TestCheckpointShowPrintsDetailsAndWrapsMissingCheckpointError(t *testing.T) {
	client := newProvenanceTestClient(t)
	created := time.Date(2026, 5, 20, 11, 12, 13, 0, time.UTC)
	saveCheckpointFixtures(t, client, []provenancepkg.Checkpoint{
		{
			ID:         "33333333-3333-3333-3333-333333333333",
			SessionKey: "session-show",
			RunID:      "run-show",
			Label:      "show me",
			Trigger:    provenancepkg.TriggerStepComplete,
			JournalSeq: 12,
			GitRef:     "abc123",
			CreatedAt:  created,
		},
	})

	cmd := NewProvenanceCmd(enabledEntBootLoader(t, client))
	out, err := executeProvenanceCommand(cmd, "checkpoint", "show", "33333333-3333-3333-3333-333333333333")

	require.NoError(t, err)
	assert.Contains(t, out, "ID:          33333333-3333-3333-3333-333333333333")
	assert.Contains(t, out, "Label:       show me")
	assert.Contains(t, out, "Trigger:     step_complete")
	assert.Contains(t, out, "Session:     session-show")
	assert.Contains(t, out, "Run:         run-show")
	assert.Contains(t, out, "Journal Seq: 12")
	assert.Contains(t, out, "Git Ref:     abc123")
	assert.Contains(t, out, "Created:     2026-05-20 11:12:13")

	missingClient := newProvenanceTestClient(t)
	cmd = NewProvenanceCmd(enabledEntBootLoader(t, missingClient))
	out, err = executeProvenanceCommand(cmd, "checkpoint", "show", "44444444-4444-4444-4444-444444444444")

	require.Error(t, err)
	assert.Contains(t, out, "Error: get checkpoint:")
	assert.Contains(t, err.Error(), "get checkpoint:")
}

func TestCheckpointCreateValidatesArgumentsBeforeBootstrap(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing run flag",
			args: []string{"checkpoint", "create", "manual-label"},
			want: `required flag(s) "run" not set`,
		},
		{
			name: "missing label",
			args: []string{"checkpoint", "create", "--run", "run-create"},
			want: "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			cmd := NewProvenanceCmd(func() (*bootstrap.Result, error) {
				calls++
				return nil, errors.New("bootstrap should not run")
			})

			out, err := executeProvenanceCommand(cmd, tt.args...)

			require.Error(t, err)
			assert.Contains(t, out, "Usage:")
			assert.Contains(t, err.Error(), tt.want)
			assert.Zero(t, calls)
		})
	}
}

func TestCheckpointCreatePrintsDisabledNoticeBeforeStoreAccess(t *testing.T) {
	cmd := NewProvenanceCmd(disabledBootLoader(t))
	out, err := executeProvenanceCommand(cmd, "checkpoint", "create", "manual-label", "--run", "run-disabled")

	require.NoError(t, err)
	assert.Contains(t, out, "Provenance is disabled")
	assert.NotContains(t, out, "runledger store unavailable")
}

func TestCheckpointCreateReturnsRunLedgerUnavailable(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Provenance.Enabled = true
	cmd := NewProvenanceCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})

	out, err := executeProvenanceCommand(cmd, "checkpoint", "create", "manual-label", "--run", "run-no-ledger")

	require.Error(t, err)
	assert.Contains(t, out, "Error: runledger store unavailable")
	assert.EqualError(t, err, "runledger store unavailable")
}

func TestCheckpointCreatePersistsManualCheckpointAndPrintsID(t *testing.T) {
	dsn := newProvenanceTestDiskDSN(t)
	client := enttest.Open(t, "sqlite3", dsn)
	ledger := runledger.NewEntStore(client)
	require.NoError(t, appendRunEvent(context.Background(), ledger, "run-create", runledger.EventRunCreated, runledger.RunCreatedPayload{
		SessionKey: "session-create",
		Goal:       "create checkpoint",
	}))
	require.NoError(t, appendRunEvent(context.Background(), ledger, "run-create", runledger.EventNoteWritten, runledger.NoteWrittenPayload{
		Key:   "note",
		Value: "value",
	}))

	cmd := NewProvenanceCmd(enabledEntBootLoader(t, client))
	out, err := executeProvenanceCommand(cmd, "checkpoint", "create", "manual-label", "--run", "run-create")

	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^Checkpoint created: [0-9a-f-]+ \(seq=2\)\n$`), out)

	queryClient := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { queryClient.Close() })
	checkpoints, err := provenancepkg.NewEntCheckpointStore(queryClient).ListByRun(context.Background(), "run-create")
	require.NoError(t, err)
	require.Len(t, checkpoints, 1)
	assert.Equal(t, "manual-label", checkpoints[0].Label)
	assert.Equal(t, provenancepkg.TriggerManual, checkpoints[0].Trigger)
	assert.Equal(t, "session-create", checkpoints[0].SessionKey)
	assert.EqualValues(t, 2, checkpoints[0].JournalSeq)
}

func newProvenanceTestClient(t *testing.T) *ent.Client {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { client.Close() })
	return client
}

func newProvenanceTestDiskDSN(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("file:%s?cache=shared&_fk=1", filepath.Join(t.TempDir(), "lango.db"))
}

func enabledEntBootLoader(t *testing.T, client *ent.Client) func() (*bootstrap.Result, error) {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Provenance.Enabled = true
	return func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:  cfg,
			Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
		}, nil
	}
}

func executeProvenanceCommand(cmd *cobra.Command, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func saveCheckpointFixtures(t *testing.T, client *ent.Client, checkpoints []provenancepkg.Checkpoint) {
	t.Helper()
	store := provenancepkg.NewEntCheckpointStore(client)
	for _, checkpoint := range checkpoints {
		require.NoError(t, store.SaveCheckpoint(context.Background(), checkpoint))
	}
}

func appendRunEvent(ctx context.Context, store runledger.RunLedgerStore, runID string, eventType runledger.JournalEventType, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   runID,
		Type:    eventType,
		Payload: data,
	})
}
