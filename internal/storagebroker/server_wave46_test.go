package storagebroker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/ent/workflowrun"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/require"
)

func TestWave46ServerRunDispatchesConfigAndRecallRequests(t *testing.T) {
	t.Parallel()

	srv := openWave11Server(t, true)
	cfg := wave11Config(t, "wave46-model")
	cfg.Agent.Provider = "wave46-provider"

	responses := runWave46Requests(t, srv, []Request{
		{ID: 4601, Method: methodHealth, DeadlineMS: 50},
		{
			ID:     4602,
			Method: methodConfigSave,
			Payload: mustPayload(t, ConfigSaveRequest{
				Name:         "wave46",
				Config:       mustRawJSON(t, cfg),
				ExplicitKeys: map[string]bool{"agent.model": true},
			}),
		},
		{
			ID:      4603,
			Method:  methodConfigExists,
			Payload: mustPayload(t, ConfigExistsRequest{Name: "wave46"}),
		},
		{
			ID:      4604,
			Method:  methodConfigSetActive,
			Payload: mustPayload(t, ConfigSetActiveRequest{Name: "wave46"}),
		},
		{ID: 4605, Method: methodConfigLoadActive},
	})

	require.Len(t, responses, 5)
	require.True(t, decodeWave46Result[HealthResult](t, responses[0]).Opened)
	require.Equal(t, map[string]bool{"saved": true}, decodeWave46Result[map[string]bool](t, responses[1]))
	require.True(t, decodeWave46Result[ConfigExistsResult](t, responses[2]).Exists)
	require.Equal(t, map[string]bool{"active": true}, decodeWave46Result[map[string]bool](t, responses[3]))

	active := decodeWave46Result[ConfigLoadActiveResult](t, responses[4])
	require.Equal(t, "wave46", active.Name)
	require.Equal(t, map[string]bool{"agent.model": true}, active.ExplicitKeys)
	var loaded config.Config
	require.NoError(t, json.Unmarshal(active.Config, &loaded))
	require.Equal(t, "wave46-provider", loaded.Agent.Provider)
	require.Equal(t, "wave46-model", loaded.Agent.Model)

	base := time.Unix(4600, 0).UTC()
	sess := session.Session{
		Key:   "wave46-session",
		Model: "wave46-test-model",
		History: []session.Message{
			{Role: types.RoleUser, Content: "remember deterministic wave forty six", Timestamp: base},
			{Role: types.RoleAssistant, Content: "recall branch indexed", Timestamp: base.Add(time.Minute)},
		},
	}
	responses = runWave46Requests(t, srv, []Request{
		{
			ID:      4610,
			Method:  methodSessionCreate,
			Payload: mustPayload(t, SessionCreateRequest{Session: sess}),
		},
		{
			ID:      4611,
			Method:  methodSessionEnd,
			Payload: mustPayload(t, SessionEndRequest{Key: sess.Key}),
		},
		{ID: 4612, Method: methodRecallProcess},
		{
			ID:      4613,
			Method:  methodRecallIndex,
			Payload: mustPayload(t, RecallIndexRequest{Key: sess.Key}),
		},
		{
			ID:      4614,
			Method:  methodRecallSearch,
			Payload: mustPayload(t, RecallSearchRequest{Query: "wave forty six", Limit: 5}),
		},
		{
			ID:      4615,
			Method:  methodRecallSummary,
			Payload: mustPayload(t, RecallSummaryRequest{Key: sess.Key}),
		},
	})

	require.Len(t, responses, 6)
	for _, resp := range responses[:4] {
		require.True(t, resp.OK, "response %d failed: %s", resp.ID, resp.Error)
		require.Empty(t, resp.Result)
	}
	search := decodeWave46Result[RecallSearchResult](t, responses[4])
	require.NotEmpty(t, search.Results)
	require.Equal(t, sess.Key, search.Results[0].RowID)
	require.NotZero(t, search.Results[0].Rank)

	summary := decodeWave46Result[RecallSummaryResult](t, responses[5])
	require.Contains(t, summary.Summary, "remember deterministic wave forty six")
}

func TestWave46ServerPaymentAndWorkflowDefaultLimitBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := openWave11Server(t, false)
	start := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	for i := 0; i < 22; i++ {
		require.NoError(t, srv.client.WorkflowRun.Create().
			SetWorkflowName(fmt.Sprintf("wave46-flow-%02d", i)).
			SetStatus(workflowrun.StatusCompleted).
			SetTotalSteps(5).
			SetCompletedSteps(i%6).
			SetStartedAt(start.Add(time.Duration(i)*time.Minute)).
			Exec(ctx))

		require.NoError(t, srv.client.PaymentTx.Create().
			SetTxHash(fmt.Sprintf("0xwave46-%02d", i)).
			SetFromAddress("0xfrom").
			SetToAddress("0xto").
			SetAmount("0.10").
			SetChainID(8453).
			SetStatus(paymenttx.StatusSubmitted).
			SetPurpose("wave46 payment").
			SetPaymentMethod(paymenttx.PaymentMethodX402V2).
			SetCreatedAt(start.Add(time.Duration(i)*time.Minute)).
			Exec(ctx))
	}
	require.NoError(t, srv.client.PaymentTx.Create().
		SetTxHash("0xwave46-invalid-amount").
		SetFromAddress("0xfrom").
		SetToAddress("0xto").
		SetAmount("not-usdc").
		SetChainID(8453).
		SetStatus(paymenttx.StatusPending).
		SetPaymentMethod(paymenttx.PaymentMethodX402V2).
		SetCreatedAt(start.Add(-2*time.Minute)).
		Exec(ctx))
	require.NoError(t, srv.client.PaymentTx.Create().
		SetTxHash("0xwave46-failed-ignored").
		SetFromAddress("0xfrom").
		SetToAddress("0xto").
		SetAmount("99.00").
		SetChainID(8453).
		SetStatus(paymenttx.StatusFailed).
		SetPaymentMethod(paymenttx.PaymentMethodX402V2).
		SetCreatedAt(start.Add(-time.Minute)).
		Exec(ctx))

	responses := runWave46Requests(t, srv, []Request{
		{ID: 4620, Method: methodWorkflowRuns, Payload: mustPayload(t, WorkflowRunsRequest{})},
		{ID: 4621, Method: methodPaymentHistory, Payload: mustPayload(t, PaymentHistoryRequest{})},
		{ID: 4622, Method: methodPaymentUsage},
	})

	workflows := decodeWave46Result[WorkflowRunsResult](t, responses[0])
	require.Len(t, workflows.Runs, 20)
	require.Equal(t, "wave46-flow-21", workflows.Runs[0].WorkflowName)
	require.Equal(t, string(workflowrun.StatusCompleted), workflows.Runs[0].Status)
	require.Equal(t, 5, workflows.Runs[0].TotalSteps)

	history := decodeWave46Result[PaymentHistoryResult](t, responses[1])
	require.Len(t, history.Entries, 20)
	require.Equal(t, "0xwave46-21", history.Entries[0].TxHash)
	require.Equal(t, string(paymenttx.StatusSubmitted), history.Entries[0].Status)
	require.Equal(t, "0.10", history.Entries[0].Amount)
	require.Equal(t, int64(8453), history.Entries[0].ChainID)

	usage := decodeWave46Result[PaymentUsageResult](t, responses[2])
	require.Equal(t, "2.20", usage.DailySpent)
}

func TestWave46ServerTargetedErrorBranches(t *testing.T) {
	t.Parallel()

	var emptyOut bytes.Buffer
	require.NoError(t, NewServer().Run(bytes.NewReader(nil), &emptyOut))
	require.Empty(t, emptyOut.String())

	srv := NewServer()
	for _, tt := range []struct {
		name    string
		req     Request
		wantErr string
	}{
		{
			name:    "open db requires path",
			req:     Request{ID: 4630, Method: methodOpenDB},
			wantErr: "open_db requires db_path",
		},
		{
			name:    "store salt requires opened database",
			req:     Request{ID: 4631, Method: methodStoreSalt, Payload: mustPayload(t, StoreSaltRequest{Salt: []byte("salt")})},
			wantErr: "database not opened",
		},
		{
			name:    "store checksum requires opened database",
			req:     Request{ID: 4632, Method: methodStoreChecksum, Payload: mustPayload(t, StoreChecksumRequest{Checksum: []byte("checksum")})},
			wantErr: "database not opened",
		},
		{
			name:    "decrypt requires payload key",
			req:     Request{ID: 4633, Method: methodDecryptPayload, Payload: mustPayload(t, DecryptPayloadRequest{})},
			wantErr: "payload protection key not initialized",
		},
		{
			name:    "config save malformed dispatch payload",
			req:     Request{ID: 4634, Method: methodConfigSave, Payload: json.RawMessage(`{"name":`)},
			wantErr: "decode broker payload",
		},
		{
			name:    "recall search malformed dispatch payload",
			req:     Request{ID: 4635, Method: methodRecallSearch, Payload: json.RawMessage(`{"query":`)},
			wantErr: "decode broker payload",
		},
		{
			name:    "workflow runs malformed dispatch payload",
			req:     Request{ID: 4636, Method: methodWorkflowRuns, Payload: json.RawMessage(`{"limit":`)},
			wantErr: "decode broker payload",
		},
		{
			name:    "payment history malformed dispatch payload",
			req:     Request{ID: 4637, Method: methodPaymentHistory, Payload: json.RawMessage(`{"limit":`)},
			wantErr: "decode broker payload",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := srv.handle(tt.req)
			require.Equal(t, tt.req.ID, resp.ID)
			require.False(t, resp.OK)
			require.Empty(t, resp.Result)
			require.Contains(t, resp.Error, tt.wantErr)
		})
	}

	noMasterKey := openWave11Server(t, false)
	resp := noMasterKey.handle(Request{ID: 4640, Method: methodConfigList})
	require.False(t, resp.OK)
	require.Contains(t, resp.Error, "master key not initialized")
}

func runWave46Requests(t *testing.T, srv *Server, requests []Request) []Response {
	t.Helper()

	var in bytes.Buffer
	for _, req := range requests {
		require.NoError(t, json.NewEncoder(&in).Encode(req))
	}
	var out bytes.Buffer
	require.NoError(t, srv.Run(&in, &out))

	dec := json.NewDecoder(&out)
	responses := make([]Response, 0, len(requests))
	for {
		var resp Response
		err := dec.Decode(&resp)
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		require.True(t, resp.OK, "response %d failed: %s", resp.ID, resp.Error)
		responses = append(responses, resp)
	}
	require.Len(t, responses, len(requests))
	return responses
}

func decodeWave46Result[T any](t *testing.T, resp Response) T {
	t.Helper()
	require.True(t, resp.OK, "response %d failed: %s", resp.ID, resp.Error)
	require.NotEmpty(t, resp.Result)

	var out T
	require.NoError(t, json.Unmarshal(resp.Result, &out))
	return out
}
