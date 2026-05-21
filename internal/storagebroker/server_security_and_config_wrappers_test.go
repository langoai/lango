package storagebroker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/auditlog"
	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/ent/workflowrun"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/require"
)

func TestServerSecurityAndConfigWrappers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := openServerSecurityAndConfigWrappersServer(t, true)

	stateAny, err := srv.dispatch(ctx, Request{Method: methodLoadSecurityState})
	require.NoError(t, err)
	state := stateAny.(LoadSecurityStateResult)
	require.True(t, state.FirstRun)
	require.Nil(t, state.Salt)
	require.Nil(t, state.Checksum)

	_, err = srv.dispatch(ctx, Request{
		Method:  methodStoreSalt,
		Payload: mustPayload(t, StoreSaltRequest{Salt: []byte("p2PToolsInventoryAndSimpleHandlers1-salt")}),
	})
	require.NoError(t, err)
	_, err = srv.dispatch(ctx, Request{
		Method:  methodStoreChecksum,
		Payload: mustPayload(t, StoreChecksumRequest{Checksum: []byte("p2PToolsInventoryAndSimpleHandlers1-checksum")}),
	})
	require.NoError(t, err)

	stateAny, err = srv.dispatch(ctx, Request{Method: methodLoadSecurityState})
	require.NoError(t, err)
	state = stateAny.(LoadSecurityStateResult)
	require.False(t, state.FirstRun)
	require.Equal(t, []byte("p2PToolsInventoryAndSimpleHandlers1-salt"), state.Salt)
	require.Equal(t, []byte("p2PToolsInventoryAndSimpleHandlers1-checksum"), state.Checksum)

	_, err = srv.dispatch(ctx, Request{
		Method: methodConfigSave,
		Payload: mustPayload(t, ConfigSaveRequest{
			Name:   "bad-json",
			Config: []byte(`{"server":`),
		}),
	})
	require.ErrorContains(t, err, "decode config payload")

	alpha := serverSecurityAndConfigWrappersConfig(t, "p2PToolsInventoryAndSimpleHandlers1-alpha")
	beta := serverSecurityAndConfigWrappersConfig(t, "p2PToolsInventoryAndSimpleHandlers1-beta")
	alphaKeys := map[string]bool{"knowledge.enabled": true}
	betaKeys := map[string]bool{"tools.exec.allowBackground": true}

	_, err = srv.dispatch(ctx, Request{
		Method: methodConfigSave,
		Payload: mustPayload(t, ConfigSaveRequest{
			Name:         "alpha",
			Config:       mustRawJSON(t, alpha),
			ExplicitKeys: alphaKeys,
		}),
	})
	require.NoError(t, err)
	_, err = srv.dispatch(ctx, Request{
		Method: methodConfigSave,
		Payload: mustPayload(t, ConfigSaveRequest{
			Name:         "beta",
			Config:       mustRawJSON(t, beta),
			ExplicitKeys: betaKeys,
		}),
	})
	require.NoError(t, err)

	existsAny, err := srv.dispatch(ctx, Request{
		Method:  methodConfigExists,
		Payload: mustPayload(t, ConfigExistsRequest{Name: "alpha"}),
	})
	require.NoError(t, err)
	require.True(t, existsAny.(ConfigExistsResult).Exists)

	_, err = srv.dispatch(ctx, Request{
		Method:  methodConfigSetActive,
		Payload: mustPayload(t, ConfigSetActiveRequest{Name: "beta"}),
	})
	require.NoError(t, err)

	activeAny, err := srv.dispatch(ctx, Request{Method: methodConfigLoadActive})
	require.NoError(t, err)
	active := activeAny.(ConfigLoadActiveResult)
	require.Equal(t, "beta", active.Name)
	require.Equal(t, betaKeys, active.ExplicitKeys)
	var activeCfg config.Config
	require.NoError(t, json.Unmarshal(active.Config, &activeCfg))
	require.Equal(t, "p2PToolsInventoryAndSimpleHandlers1-beta", activeCfg.Agent.Model)

	loadedAny, err := srv.dispatch(ctx, Request{
		Method:  methodConfigLoad,
		Payload: mustPayload(t, ConfigLoadRequest{Name: "alpha"}),
	})
	require.NoError(t, err)
	loaded := loadedAny.(ConfigLoadResult)
	require.Equal(t, alphaKeys, loaded.ExplicitKeys)
	var loadedCfg config.Config
	require.NoError(t, json.Unmarshal(loaded.Config, &loadedCfg))
	require.Equal(t, "p2PToolsInventoryAndSimpleHandlers1-alpha", loadedCfg.Agent.Model)

	listAny, err := srv.dispatch(ctx, Request{Method: methodConfigList})
	require.NoError(t, err)
	list := listAny.(ConfigListResult)
	require.Len(t, list.Profiles, 2)
	require.Equal(t, "alpha", list.Profiles[0].Name)
	require.False(t, list.Profiles[0].Active)
	require.Equal(t, "beta", list.Profiles[1].Name)
	require.True(t, list.Profiles[1].Active)
	require.NotEmpty(t, list.Profiles[0].CreatedAt)
	require.NotEmpty(t, list.Profiles[0].UpdatedAt)

	_, err = srv.dispatch(ctx, Request{
		Method:  methodConfigDelete,
		Payload: mustPayload(t, ConfigDeleteRequest{Name: "alpha"}),
	})
	require.NoError(t, err)

	existsAny, err = srv.dispatch(ctx, Request{
		Method:  methodConfigExists,
		Payload: mustPayload(t, ConfigExistsRequest{Name: "alpha"}),
	})
	require.NoError(t, err)
	require.False(t, existsAny.(ConfigExistsResult).Exists)
}

func TestServerSessionRecallAndOperationalWrappers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := openServerSecurityAndConfigWrappersServer(t, false)
	base := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	yesterday := time.Now().AddDate(0, 0, -1).UTC().Truncate(time.Second)
	currentUsageAt := time.Now().Add(time.Hour).Truncate(time.Second)

	sess := session.Session{
		Key:   "p2PToolsInventoryAndSimpleHandlers1-session",
		Model: "test-model",
		History: []session.Message{
			{Role: types.RoleUser, Content: "remember broker recall", Timestamp: base},
			{Role: types.RoleAssistant, Content: "indexed reply", Timestamp: base.Add(time.Minute)},
		},
	}
	_, err := srv.dispatch(ctx, Request{
		Method:  methodSessionCreate,
		Payload: mustPayload(t, SessionCreateRequest{Session: sess}),
	})
	require.NoError(t, err)
	sess.Model = "updated-model"
	_, err = srv.dispatch(ctx, Request{
		Method:  methodSessionUpdate,
		Payload: mustPayload(t, SessionUpdateRequest{Session: sess}),
	})
	require.NoError(t, err)
	_, err = srv.dispatch(ctx, Request{
		Method:  methodSessionSetSalt,
		Payload: mustPayload(t, SessionSetSaltRequest{Name: "p2PToolsInventoryAndSimpleHandlers1-agent", Salt: []byte("agent-salt")}),
	})
	require.NoError(t, err)

	saltAny, err := srv.dispatch(ctx, Request{
		Method:  methodSessionGetSalt,
		Payload: mustPayload(t, SessionGetSaltRequest{Name: "p2PToolsInventoryAndSimpleHandlers1-agent"}),
	})
	require.NoError(t, err)
	require.Equal(t, []byte("agent-salt"), saltAny.(SessionGetSaltResult).Salt)

	_, err = srv.dispatch(ctx, Request{
		Method:  methodSessionEnd,
		Payload: mustPayload(t, SessionEndRequest{Key: "p2PToolsInventoryAndSimpleHandlers1-session"}),
	})
	require.NoError(t, err)
	_, err = srv.dispatch(ctx, Request{Method: methodRecallProcess})
	require.NoError(t, err)
	_, err = srv.dispatch(ctx, Request{
		Method:  methodRecallIndex,
		Payload: mustPayload(t, RecallIndexRequest{Key: "p2PToolsInventoryAndSimpleHandlers1-session"}),
	})
	require.NoError(t, err)

	searchAny, err := srv.dispatch(ctx, Request{
		Method:  methodRecallSearch,
		Payload: mustPayload(t, RecallSearchRequest{Query: "broker recall", Limit: 3}),
	})
	require.NoError(t, err)
	require.NotEmpty(t, searchAny.(RecallSearchResult).Results)

	summaryAny, err := srv.dispatch(ctx, Request{
		Method:  methodRecallSummary,
		Payload: mustPayload(t, RecallSummaryRequest{Key: "p2PToolsInventoryAndSimpleHandlers1-session"}),
	})
	require.NoError(t, err)
	require.Contains(t, summaryAny.(RecallSummaryResult).Summary, "remember broker recall")

	require.NoError(t, srv.client.WorkflowRun.Create().
		SetWorkflowName("old-flow").
		SetStatus(workflowrun.StatusRunning).
		SetTotalSteps(3).
		SetCompletedSteps(1).
		SetStartedAt(base).
		Exec(ctx))
	require.NoError(t, srv.client.WorkflowRun.Create().
		SetWorkflowName("new-flow").
		SetStatus(workflowrun.StatusCompleted).
		SetTotalSteps(4).
		SetCompletedSteps(4).
		SetStartedAt(base.Add(time.Hour)).
		Exec(ctx))
	require.NoError(t, srv.client.AuditLog.Create().
		SetAction(auditlog.ActionAlert).
		SetActor("watcher").
		SetTarget("quota").
		SetDetails(map[string]interface{}{"severity": "high"}).
		SetTimestamp(base.Add(time.Hour)).
		Exec(ctx))
	require.NoError(t, srv.client.AuditLog.Create().
		SetAction(auditlog.ActionToolCall).
		SetActor("worker").
		SetTarget("ignored").
		SetTimestamp(base.Add(time.Hour)).
		Exec(ctx))
	require.NoError(t, srv.client.PeerReputation.Create().
		SetPeerDid("did:example:p2PToolsInventoryAndSimpleHandlers1").
		SetSuccessfulExchanges(3).
		SetFailedExchanges(1).
		SetTimeoutCount(2).
		SetTrustScore(0.42).
		SetFirstSeen(base).
		SetLastInteraction(base.Add(time.Hour)).
		Exec(ctx))
	require.NoError(t, srv.client.PaymentTx.Create().
		SetTxHash("0xold").
		SetFromAddress("0xfrom").
		SetToAddress("0xto").
		SetAmount("1.25").
		SetChainID(8453).
		SetStatus(paymenttx.StatusConfirmed).
		SetPurpose("old payment").
		SetX402URL("https://example.invalid/old").
		SetPaymentMethod(paymenttx.PaymentMethodX402V2).
		SetCreatedAt(yesterday).
		Exec(ctx))
	require.NoError(t, srv.client.PaymentTx.Create().
		SetTxHash("0xnew").
		SetFromAddress("0xfrom2").
		SetToAddress("0xto2").
		SetAmount("2.00").
		SetChainID(8453).
		SetStatus(paymenttx.StatusSubmitted).
		SetPurpose("new payment").
		SetErrorMessage("pending confirmation").
		SetCreatedAt(currentUsageAt).
		Exec(ctx))

	workflowsAny, err := srv.dispatch(ctx, Request{
		Method:  methodWorkflowRuns,
		Payload: mustPayload(t, WorkflowRunsRequest{Limit: 1}),
	})
	require.NoError(t, err)
	workflows := workflowsAny.(WorkflowRunsResult)
	require.Len(t, workflows.Runs, 1)
	require.Equal(t, "new-flow", workflows.Runs[0].WorkflowName)
	require.Equal(t, string(workflowrun.StatusCompleted), workflows.Runs[0].Status)

	alertsAny, err := srv.dispatch(ctx, Request{
		Method:  methodAlerts,
		Payload: mustPayload(t, AlertsRequest{From: base.Add(30 * time.Minute)}),
	})
	require.NoError(t, err)
	alerts := alertsAny.(AlertsResult)
	require.Len(t, alerts.Alerts, 1)
	require.Equal(t, "quota", alerts.Alerts[0].Type)
	require.Equal(t, "watcher", alerts.Alerts[0].Actor)
	require.Equal(t, "high", alerts.Alerts[0].Details["severity"])

	missingRepAny, err := srv.dispatch(ctx, Request{
		Method:  methodReputationGet,
		Payload: mustPayload(t, ReputationGetRequest{PeerDID: "did:example:missing"}),
	})
	require.NoError(t, err)
	missingRep := missingRepAny.(ReputationGetResult)
	require.False(t, missingRep.Found)
	require.Equal(t, "did:example:missing", missingRep.PeerDID)

	repAny, err := srv.dispatch(ctx, Request{
		Method:  methodReputationGet,
		Payload: mustPayload(t, ReputationGetRequest{PeerDID: "did:example:p2PToolsInventoryAndSimpleHandlers1"}),
	})
	require.NoError(t, err)
	rep := repAny.(ReputationGetResult)
	require.True(t, rep.Found)
	require.Equal(t, 0.42, rep.TrustScore)
	require.Equal(t, 3, rep.SuccessfulExchanges)
	require.Equal(t, 1, rep.FailedExchanges)
	require.Equal(t, 2, rep.TimeoutCount)

	historyAny, err := srv.dispatch(ctx, Request{
		Method:  methodPaymentHistory,
		Payload: mustPayload(t, PaymentHistoryRequest{Limit: 1}),
	})
	require.NoError(t, err)
	history := historyAny.(PaymentHistoryResult)
	require.Len(t, history.Entries, 1)
	require.Equal(t, "0xnew", history.Entries[0].TxHash)
	require.Equal(t, string(paymenttx.StatusSubmitted), history.Entries[0].Status)
	require.Equal(t, "pending confirmation", history.Entries[0].ErrorMessage)

	usageAny, err := srv.dispatch(ctx, Request{Method: methodPaymentUsage})
	require.NoError(t, err)
	require.Equal(t, "2.00", usageAny.(PaymentUsageResult).DailySpent)
}

func TestClientWrappersThroughFakeBroker(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Unix(200, 0).UTC()

	expectClientCall(t, func(c *Client) error {
		got, err := c.Health(ctx)
		require.Equal(t, HealthResult{Opened: true}, got)
		return err
	}, methodHealth, nil, HealthResult{Opened: true})

	expectClientCall(t, func(c *Client) error {
		got, err := c.OpenDB(ctx, OpenDBRequest{DBPath: "db.sqlite"})
		require.Equal(t, OpenDBResult{Opened: true}, got)
		return err
	}, methodOpenDB, OpenDBRequest{DBPath: "db.sqlite"}, OpenDBResult{Opened: true})

	expectClientCall(t, func(c *Client) error {
		got, err := c.DBStatusSummary(ctx, DBStatusSummaryRequest{DBPath: "db.sqlite"})
		require.Equal(t, DBStatusSummaryResult{Available: true, EncryptionKeys: 2, StoredSecrets: 3}, got)
		return err
	}, methodDBStatus, DBStatusSummaryRequest{DBPath: "db.sqlite"}, DBStatusSummaryResult{Available: true, EncryptionKeys: 2, StoredSecrets: 3})

	expectClientCall(t, func(c *Client) error {
		got, err := c.DecryptPayload(ctx, []byte("cipher"), []byte("nonce"), 4)
		require.Equal(t, []byte("plain"), got.Plaintext)
		return err
	}, methodDecryptPayload, DecryptPayloadRequest{Ciphertext: []byte("cipher"), Nonce: []byte("nonce"), KeyVersion: 4}, DecryptPayloadResult{Plaintext: []byte("plain")})

	expectClientCall(t, func(c *Client) error {
		got, err := c.LoadSecurityState(ctx)
		require.Equal(t, LoadSecurityStateResult{Salt: []byte("salt"), Checksum: []byte("sum"), FirstRun: false}, got)
		return err
	}, methodLoadSecurityState, LoadSecurityStateRequest{}, LoadSecurityStateResult{Salt: []byte("salt"), Checksum: []byte("sum"), FirstRun: false})

	expectClientCall(t, func(c *Client) error {
		got, err := c.ConfigLoad(ctx, "alpha")
		require.Equal(t, []byte(`{"ok":true}`), got.Config)
		require.True(t, got.ExplicitKeys["knowledge.enabled"])
		return err
	}, methodConfigLoad, ConfigLoadRequest{Name: "alpha"}, ConfigLoadResult{Config: []byte(`{"ok":true}`), ExplicitKeys: map[string]bool{"knowledge.enabled": true}})

	expectClientCall(t, func(c *Client) error {
		got, err := c.ConfigLoadActive(ctx)
		require.Equal(t, "beta", got.Name)
		return err
	}, methodConfigLoadActive, nil, ConfigLoadActiveResult{Name: "beta", Config: []byte(`{"active":true}`)})

	saveCfg := struct {
		Model string `json:"model"`
	}{Model: "p2PToolsInventoryAndSimpleHandlers1"}
	expectClientCall(t, func(c *Client) error {
		return c.ConfigSave(ctx, "alpha", saveCfg, map[string]bool{"agent.model": true})
	}, methodConfigSave, ConfigSaveRequest{Name: "alpha", Config: mustRawJSON(t, saveCfg), ExplicitKeys: map[string]bool{"agent.model": true}}, nil)

	expectClientCall(t, func(c *Client) error {
		return c.ConfigSetActive(ctx, "alpha")
	}, methodConfigSetActive, ConfigSetActiveRequest{Name: "alpha"}, nil)

	expectClientCall(t, func(c *Client) error {
		got, err := c.ConfigList(ctx)
		require.Len(t, got.Profiles, 1)
		require.Equal(t, "alpha", got.Profiles[0].Name)
		return err
	}, methodConfigList, nil, ConfigListResult{Profiles: []ConfigProfileInfo{{Name: "alpha", Active: true, Version: 2}}})

	expectClientCall(t, func(c *Client) error {
		return c.ConfigDelete(ctx, "alpha")
	}, methodConfigDelete, ConfigDeleteRequest{Name: "alpha"}, nil)

	expectClientCall(t, func(c *Client) error {
		got, err := c.ConfigExists(ctx, "alpha")
		require.True(t, got.Exists)
		return err
	}, methodConfigExists, ConfigExistsRequest{Name: "alpha"}, ConfigExistsResult{Exists: true})

	sess := &session.Session{Key: "s1", Model: "model"}
	expectClientCall(t, func(c *Client) error {
		return c.SessionCreate(ctx, sess)
	}, methodSessionCreate, SessionCreateRequest{Session: *sess}, nil)
	expectClientCall(t, func(c *Client) error {
		got, err := c.SessionGet(ctx, "s1")
		require.Equal(t, "s1", got.Key)
		return err
	}, methodSessionGet, SessionGetRequest{Key: "s1"}, SessionGetResult{Session: sess})
	expectClientCall(t, func(c *Client) error {
		return c.SessionUpdate(ctx, sess)
	}, methodSessionUpdate, SessionUpdateRequest{Session: *sess}, nil)
	expectClientCall(t, func(c *Client) error {
		return c.SessionDelete(ctx, "s1")
	}, methodSessionDelete, SessionDeleteRequest{Key: "s1"}, nil)
	msg := session.Message{Role: types.RoleUser, Content: "hello", Timestamp: now}
	expectClientCall(t, func(c *Client) error {
		return c.SessionAppendMessage(ctx, "s1", msg)
	}, methodSessionAppend, SessionAppendMessageRequest{Key: "s1", Message: msg}, nil)
	expectClientCall(t, func(c *Client) error {
		return c.SessionEnd(ctx, "s1")
	}, methodSessionEnd, SessionEndRequest{Key: "s1"}, nil)
	expectClientCall(t, func(c *Client) error {
		got, err := c.SessionList(ctx)
		require.Equal(t, []session.SessionSummary{{Key: "s1", CreatedAt: now, UpdatedAt: now}}, got)
		return err
	}, methodSessionList, nil, SessionListResult{Sessions: []SessionSummaryRecord{{Key: "s1", CreatedAt: now, UpdatedAt: now}}})
	expectClientCall(t, func(c *Client) error {
		got, err := c.SessionGetSalt(ctx, "agent")
		require.Equal(t, []byte("salt"), got)
		return err
	}, methodSessionGetSalt, SessionGetSaltRequest{Name: "agent"}, SessionGetSaltResult{Salt: []byte("salt")})
	expectClientCall(t, func(c *Client) error {
		return c.SessionSetSalt(ctx, "agent", []byte("salt"))
	}, methodSessionSetSalt, SessionSetSaltRequest{Name: "agent", Salt: []byte("salt")}, nil)

	expectClientCall(t, func(c *Client) error {
		return c.RecallIndexSession(ctx, "s1")
	}, methodRecallIndex, RecallIndexRequest{Key: "s1"}, nil)
	expectClientCall(t, func(c *Client) error {
		return c.RecallProcessPending(ctx)
	}, methodRecallProcess, nil, nil)
	expectClientCall(t, func(c *Client) error {
		got, err := c.RecallSearch(ctx, "hello", 2)
		require.Equal(t, []search.SearchResult{{RowID: "s1", Rank: 0.5}}, got)
		return err
	}, methodRecallSearch, RecallSearchRequest{Query: "hello", Limit: 2}, RecallSearchResult{Results: []RecallSearchRecord{{RowID: "s1", Rank: 0.5}}})
	expectClientCall(t, func(c *Client) error {
		got, err := c.RecallGetSummary(ctx, "s1")
		require.Equal(t, "summary", got)
		return err
	}, methodRecallSummary, RecallSummaryRequest{Key: "s1"}, RecallSummaryResult{Summary: "summary"})

	expectClientCall(t, func(c *Client) error {
		got, err := c.LearningHistory(ctx, 7)
		require.Len(t, got.Entries, 1)
		return err
	}, methodLearningHistory, LearningHistoryRequest{Limit: 7}, LearningHistoryResult{Entries: []LearningHistoryRecord{{ID: "learn-1", Trigger: "trigger"}}})
	expectClientCall(t, func(c *Client) error {
		got, err := c.PendingInquiries(ctx, 8)
		require.Len(t, got.Entries, 1)
		return err
	}, methodPendingInquiries, PendingInquiriesRequest{Limit: 8}, PendingInquiriesResult{Entries: []PendingInquiryRecord{{ID: "inq-1", Topic: "topic"}}})
	expectClientCall(t, func(c *Client) error {
		got, err := c.WorkflowRuns(ctx, 9)
		require.Len(t, got.Runs, 1)
		return err
	}, methodWorkflowRuns, WorkflowRunsRequest{Limit: 9}, WorkflowRunsResult{Runs: []WorkflowRunRecord{{WorkflowName: "flow"}}})
	expectClientCall(t, func(c *Client) error {
		got, err := c.Alerts(ctx, now)
		require.Len(t, got.Alerts, 1)
		return err
	}, methodAlerts, AlertsRequest{From: now}, AlertsResult{Alerts: []AlertRecord{{Type: "quota", Actor: "watcher"}}})
	expectClientCall(t, func(c *Client) error {
		got, err := c.ReputationGet(ctx, "did:example:peer")
		require.True(t, got.Found)
		return err
	}, methodReputationGet, ReputationGetRequest{PeerDID: "did:example:peer"}, ReputationGetResult{PeerDID: "did:example:peer", Found: true})
	expectClientCall(t, func(c *Client) error {
		got, err := c.PaymentHistory(ctx, 10)
		require.Len(t, got.Entries, 1)
		return err
	}, methodPaymentHistory, PaymentHistoryRequest{Limit: 10}, PaymentHistoryResult{Entries: []PaymentHistoryRecord{{TxHash: "0xabc"}}})
	expectClientCall(t, func(c *Client) error {
		got, err := c.PaymentUsage(ctx)
		require.Equal(t, "1.00", got.DailySpent)
		return err
	}, methodPaymentUsage, nil, PaymentUsageResult{DailySpent: "1.00"})
}

func TestClientResultWrappersReturnZeroValuesOnBrokerError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	wantErr := "broker unavailable"
	now := time.Unix(300, 0).UTC()

	tests := []struct {
		name string
		run  func(*testing.T, *Client) error
	}{
		{
			name: "health",
			run: func(t *testing.T, c *Client) error {
				got, err := c.Health(ctx)
				require.Equal(t, HealthResult{}, got)
				return err
			},
		},
		{
			name: "open db",
			run: func(t *testing.T, c *Client) error {
				got, err := c.OpenDB(ctx, OpenDBRequest{DBPath: "db.sqlite"})
				require.Equal(t, OpenDBResult{}, got)
				return err
			},
		},
		{
			name: "db status summary",
			run: func(t *testing.T, c *Client) error {
				got, err := c.DBStatusSummary(ctx, DBStatusSummaryRequest{DBPath: "db.sqlite"})
				require.Equal(t, DBStatusSummaryResult{}, got)
				return err
			},
		},
		{
			name: "encrypt payload",
			run: func(t *testing.T, c *Client) error {
				got, err := c.EncryptPayload(ctx, []byte("plain"))
				require.Equal(t, EncryptPayloadResult{}, got)
				return err
			},
		},
		{
			name: "decrypt payload",
			run: func(t *testing.T, c *Client) error {
				got, err := c.DecryptPayload(ctx, []byte("cipher"), []byte("nonce"), 4)
				require.Equal(t, DecryptPayloadResult{}, got)
				return err
			},
		},
		{
			name: "load security state",
			run: func(t *testing.T, c *Client) error {
				got, err := c.LoadSecurityState(ctx)
				require.Equal(t, LoadSecurityStateResult{}, got)
				return err
			},
		},
		{
			name: "config load",
			run: func(t *testing.T, c *Client) error {
				got, err := c.ConfigLoad(ctx, "alpha")
				require.Equal(t, ConfigLoadResult{}, got)
				return err
			},
		},
		{
			name: "config load active",
			run: func(t *testing.T, c *Client) error {
				got, err := c.ConfigLoadActive(ctx)
				require.Equal(t, ConfigLoadActiveResult{}, got)
				return err
			},
		},
		{
			name: "config list",
			run: func(t *testing.T, c *Client) error {
				got, err := c.ConfigList(ctx)
				require.Equal(t, ConfigListResult{}, got)
				return err
			},
		},
		{
			name: "config exists",
			run: func(t *testing.T, c *Client) error {
				got, err := c.ConfigExists(ctx, "alpha")
				require.Equal(t, ConfigExistsResult{}, got)
				return err
			},
		},
		{
			name: "session get",
			run: func(t *testing.T, c *Client) error {
				got, err := c.SessionGet(ctx, "s1")
				require.Nil(t, got)
				return err
			},
		},
		{
			name: "session list",
			run: func(t *testing.T, c *Client) error {
				got, err := c.SessionList(ctx)
				require.Nil(t, got)
				return err
			},
		},
		{
			name: "session get salt",
			run: func(t *testing.T, c *Client) error {
				got, err := c.SessionGetSalt(ctx, "agent")
				require.Nil(t, got)
				return err
			},
		},
		{
			name: "recall search",
			run: func(t *testing.T, c *Client) error {
				got, err := c.RecallSearch(ctx, "hello", 2)
				require.Nil(t, got)
				return err
			},
		},
		{
			name: "recall summary",
			run: func(t *testing.T, c *Client) error {
				got, err := c.RecallGetSummary(ctx, "s1")
				require.Empty(t, got)
				return err
			},
		},
		{
			name: "learning history",
			run: func(t *testing.T, c *Client) error {
				got, err := c.LearningHistory(ctx, 7)
				require.Equal(t, LearningHistoryResult{}, got)
				return err
			},
		},
		{
			name: "pending inquiries",
			run: func(t *testing.T, c *Client) error {
				got, err := c.PendingInquiries(ctx, 8)
				require.Equal(t, PendingInquiriesResult{}, got)
				return err
			},
		},
		{
			name: "workflow runs",
			run: func(t *testing.T, c *Client) error {
				got, err := c.WorkflowRuns(ctx, 9)
				require.Equal(t, WorkflowRunsResult{}, got)
				return err
			},
		},
		{
			name: "alerts",
			run: func(t *testing.T, c *Client) error {
				got, err := c.Alerts(ctx, now)
				require.Equal(t, AlertsResult{}, got)
				return err
			},
		},
		{
			name: "reputation get",
			run: func(t *testing.T, c *Client) error {
				got, err := c.ReputationGet(ctx, "did:example:peer")
				require.Equal(t, ReputationGetResult{}, got)
				return err
			},
		},
		{
			name: "payment history",
			run: func(t *testing.T, c *Client) error {
				got, err := c.PaymentHistory(ctx, 10)
				require.Equal(t, PaymentHistoryResult{}, got)
				return err
			},
		},
		{
			name: "payment usage",
			run: func(t *testing.T, c *Client) error {
				got, err := c.PaymentUsage(ctx)
				require.Equal(t, PaymentUsageResult{}, got)
				return err
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, cleanup := newPipeClient(t, func(req Request) Response {
				return Response{ID: req.ID, OK: false, Error: wantErr}
			})
			defer cleanup()

			require.ErrorContains(t, tt.run(t, c), wantErr)
		})
	}
}

func TestClientCallErrorBranches(t *testing.T) {
	t.Parallel()

	closed := &Client{closed: true}
	_, err := closed.Health(context.Background())
	require.ErrorContains(t, err, "broker client closed")

	stdin := &bufferWriteCloser{}
	c := &Client{
		stdin:   stdin,
		stdout:  io.NopCloser(bytes.NewReader(nil)),
		pending: make(map[uint64]chan Response),
	}
	err = c.ConfigSave(context.Background(), "bad", func() {}, nil)
	require.ErrorContains(t, err, "marshal config payload")
	require.Empty(t, stdin.String())
	require.Empty(t, c.pending)

	writeErrClient := &Client{
		stdin:   errWriteCloser{},
		stdout:  io.NopCloser(bytes.NewReader(nil)),
		pending: make(map[uint64]chan Response),
	}
	err = writeErrClient.StoreSalt(context.Background(), []byte("salt"))
	require.ErrorContains(t, err, "write broker request")
	require.Empty(t, writeErrClient.pending)

	respR, respW := io.Pipe()
	closedResponseClient := &Client{
		stdin:   &closeOnWriteCloser{close: func() { _ = respW.Close() }},
		stdout:  respR,
		pending: make(map[uint64]chan Response),
	}
	go closedResponseClient.readLoop()
	err = closedResponseClient.StoreSalt(context.Background(), []byte("salt"))
	require.ErrorContains(t, err, "broker response channel closed")

	closeClient := &Client{
		stdin:   &errCloseWriteCloser{},
		stdout:  io.NopCloser(bytes.NewReader(nil)),
		pending: make(map[uint64]chan Response),
	}
	closeCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	err = closeClient.Close(closeCtx)
	require.ErrorContains(t, err, "close failed")
	require.True(t, closeClient.closed)
}

func TestClientMaxReturnsSecondWhenGreaterOrEqual(t *testing.T) {
	t.Parallel()

	require.Equal(t, int64(7), max(7, 5))
	require.Equal(t, int64(5), max(3, 5))
	require.Equal(t, int64(5), max(5, 5))
}

func openServerSecurityAndConfigWrappersServer(t *testing.T, withMasterKey bool) *Server {
	t.Helper()

	srv := NewServer()
	req := OpenDBRequest{DBPath: t.TempDir() + "/broker.db"}
	if withMasterKey {
		req.MasterKey = bytes.Repeat([]byte{0x4d}, 32)
	}
	_, err := srv.dispatch(context.Background(), Request{
		Method:  methodOpenDB,
		Payload: mustPayload(t, req),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = srv.shutdown()
	})
	return srv
}

func serverSecurityAndConfigWrappersConfig(t *testing.T, model string) *config.Config {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Agent.Provider = ""
	cfg.Agent.Model = model
	cfg.Session.DatabasePath = "sessions.db"
	cfg.Graph.DatabasePath = "graph.db"
	cfg.Skill.SkillsDir = "skills"
	cfg.Workflow.StateDir = "workflow"
	cfg.P2P.KeyDir = "p2p/keys"
	cfg.P2P.ZKP.ProofCacheDir = "p2p/proofs"
	cfg.P2P.Workspace.DataDir = "p2p/workspace"
	return cfg
}

func mustRawJSON(t *testing.T, v any) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func expectClientCall(t *testing.T, call func(*Client) error, method string, payload any, result any) {
	t.Helper()

	seen := make(chan Request, 1)
	c, cleanup := newPipeClient(t, func(req Request) Response {
		seen <- req
		resp := Response{ID: req.ID, OK: true}
		if result != nil {
			resp.Result = mustPayload(t, result)
		}
		return resp
	})
	defer cleanup()

	require.NoError(t, call(c))
	req := <-seen
	require.Equal(t, method, req.Method)
	if payload == nil {
		require.Empty(t, req.Payload)
		return
	}
	require.JSONEq(t, string(mustRawJSON(t, payload)), string(req.Payload))
}

type errWriteCloser struct{}

func (errWriteCloser) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (errWriteCloser) Close() error {
	return nil
}

type errCloseWriteCloser struct {
	bufferWriteCloser
}

func (errCloseWriteCloser) Close() error {
	return errors.New("close failed")
}

type closeOnWriteCloser struct {
	bytes.Buffer
	once  sync.Once
	close func()
}

func (w *closeOnWriteCloser) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	w.once.Do(w.close)
	return n, err
}

func (w *closeOnWriteCloser) Close() error {
	w.once.Do(w.close)
	return nil
}
