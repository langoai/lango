package firewall

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestWave41ACLActionValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		give ACLAction
		want bool
	}{
		{name: "allow is valid", give: ACLActionAllow, want: true},
		{name: "deny is valid", give: ACLActionDeny, want: true},
		{name: "empty action is invalid", give: ACLAction(""), want: false},
		{name: "unknown action is invalid", give: ACLAction("audit"), want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.give.Valid())
		})
	}
}

func TestWave41ValidateRuleAllowAndDenyCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rule    ACLRule
		wantErr bool
	}{
		{
			name:    "allow wildcard peer with empty tools is rejected",
			rule:    ACLRule{PeerDID: WildcardAll, Action: ACLActionAllow},
			wantErr: true,
		},
		{
			name: "allow wildcard peer with wildcard tool is rejected",
			rule: ACLRule{
				PeerDID: WildcardAll,
				Action:  ACLActionAllow,
				Tools:   []string{WildcardAll},
			},
			wantErr: true,
		},
		{
			name: "allow wildcard peer with specific tool is accepted",
			rule: ACLRule{
				PeerDID: WildcardAll,
				Action:  ACLActionAllow,
				Tools:   []string{"search"},
			},
		},
		{
			name: "allow specific peer with all tools is accepted",
			rule: ACLRule{
				PeerDID: "did:key:peer-1",
				Action:  ACLActionAllow,
			},
		},
		{
			name: "deny wildcard peer with wildcard tool is accepted",
			rule: ACLRule{
				PeerDID: WildcardAll,
				Action:  ACLActionDeny,
				Tools:   []string{WildcardAll},
			},
		},
		{
			name: "unknown non-allow action is accepted by safety validator",
			rule: ACLRule{
				PeerDID: WildcardAll,
				Action:  ACLAction("audit"),
				Tools:   []string{WildcardAll},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateRule(tt.rule)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "overly permissive")
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestWave41NewLoadsRulesAndInitializesRateLimiters(t *testing.T) {
	t.Parallel()

	rules := []ACLRule{
		{PeerDID: "did:key:limited", Action: ACLActionAllow, Tools: []string{"echo"}, RateLimit: 2},
		{PeerDID: WildcardAll, Action: ACLActionDeny, Tools: []string{"blocked"}, RateLimit: 1},
		{PeerDID: "did:key:unlimited", Action: ACLActionAllow, Tools: []string{"echo"}},
		{Action: ACLActionDeny, Tools: []string{"anonymous"}, RateLimit: 3},
	}

	fw := New(rules, testLogger())

	require.Len(t, fw.Rules(), len(rules))
	assert.Contains(t, fw.limiters, "did:key:limited")
	assert.Contains(t, fw.limiters, WildcardAll)
	assert.NotContains(t, fw.limiters, "did:key:unlimited")
	assert.NotContains(t, fw.limiters, "")
}

func TestWave41AddRemoveAndRulesCopySemantics(t *testing.T) {
	t.Parallel()

	fw := New(nil, testLogger())

	err := fw.AddRule(ACLRule{
		PeerDID:   "did:key:peer-1",
		Action:    ACLActionAllow,
		Tools:     []string{"echo"},
		RateLimit: 1,
	})
	require.NoError(t, err)
	require.Contains(t, fw.limiters, "did:key:peer-1")

	err = fw.AddRule(ACLRule{
		PeerDID: "did:key:peer-1",
		Action:  ACLActionDeny,
		Tools:   []string{"admin"},
	})
	require.NoError(t, err)
	require.Len(t, fw.Rules(), 2)

	copied := fw.Rules()
	copied[0].PeerDID = "did:key:mutated"
	copied[0].Tools[0] = "mutated"
	copied = append(copied, ACLRule{PeerDID: "did:key:extra", Action: ACLActionDeny})

	stored := fw.Rules()
	require.Len(t, stored, 2)
	assert.Equal(t, "did:key:peer-1", stored[0].PeerDID)
	assert.Equal(t, []string{"echo"}, stored[0].Tools)

	removed := fw.RemoveRule("did:key:peer-1")
	assert.Equal(t, 2, removed)
	assert.Empty(t, fw.Rules())
	assert.NotContains(t, fw.limiters, "did:key:peer-1")

	assert.Zero(t, fw.RemoveRule("did:key:missing"))
}

func TestWave41RuleInputsAreCopiedOnNewAndAddRule(t *testing.T) {
	t.Parallel()

	initialTools := []string{"echo"}
	fw := New([]ACLRule{{
		PeerDID: WildcardAll,
		Action:  ACLActionAllow,
		Tools:   initialTools,
	}}, testLogger())
	initialTools[0] = WildcardAll

	err := fw.FilterQuery(context.Background(), "did:key:any", "admin")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoMatchingAllowRule)

	addedTools := []string{"search*"}
	require.NoError(t, fw.AddRule(ACLRule{
		PeerDID: "did:key:peer-1",
		Action:  ACLActionAllow,
		Tools:   addedTools,
	}))
	addedTools[0] = WildcardAll

	err = fw.FilterQuery(context.Background(), "did:key:peer-1", "admin")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoMatchingAllowRule)
}

func TestWave41FilterQueryACLBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rules   []ACLRule
		peerDID string
		tool    string
		wantErr error
	}{
		{
			name: "matching allow permits query",
			rules: []ACLRule{
				{PeerDID: "did:key:peer-1", Action: ACLActionAllow, Tools: []string{"echo"}},
			},
			peerDID: "did:key:peer-1",
			tool:    "echo",
		},
		{
			name: "matching deny overrides allow",
			rules: []ACLRule{
				{PeerDID: "did:key:peer-1", Action: ACLActionAllow, Tools: []string{"echo"}},
				{PeerDID: "did:key:peer-1", Action: ACLActionDeny, Tools: []string{"echo"}},
			},
			peerDID: "did:key:peer-1",
			tool:    "echo",
			wantErr: ErrQueryDenied,
		},
		{
			name: "missing allow denies query",
			rules: []ACLRule{
				{PeerDID: "did:key:peer-1", Action: ACLActionAllow, Tools: []string{"echo"}},
			},
			peerDID: "did:key:peer-2",
			tool:    "echo",
			wantErr: ErrNoMatchingAllowRule,
		},
		{
			name: "tool mismatch denies query",
			rules: []ACLRule{
				{PeerDID: "did:key:peer-1", Action: ACLActionAllow, Tools: []string{"echo"}},
			},
			peerDID: "did:key:peer-1",
			tool:    "admin",
			wantErr: ErrNoMatchingAllowRule,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fw := New(tt.rules, testLogger())
			err := fw.FilterQuery(context.Background(), tt.peerDID, tt.tool)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestWave41FilterQueryRateLimitBranches(t *testing.T) {
	t.Parallel()

	t.Run("peer rate limiter denies second request", func(t *testing.T) {
		t.Parallel()

		fw := New([]ACLRule{
			{
				PeerDID:   "did:key:limited",
				Action:    ACLActionAllow,
				Tools:     []string{"echo"},
				RateLimit: 1,
			},
		}, testLogger())
		fw.limiters["did:key:limited"] = rate.NewLimiter(0, 1)

		require.NoError(t, fw.FilterQuery(context.Background(), "did:key:limited", "echo"))
		err := fw.FilterQuery(context.Background(), "did:key:limited", "echo")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrRateLimitExceeded)
	})

	t.Run("global rate limiter denies second request", func(t *testing.T) {
		t.Parallel()

		fw := New([]ACLRule{
			{
				PeerDID:   WildcardAll,
				Action:    ACLActionDeny,
				Tools:     []string{"blocked"},
				RateLimit: 1,
			},
			{PeerDID: "did:key:peer-1", Action: ACLActionAllow, Tools: []string{"echo"}},
		}, testLogger())
		fw.limiters[WildcardAll] = rate.NewLimiter(0, 1)

		require.NoError(t, fw.FilterQuery(context.Background(), "did:key:peer-1", "echo"))
		err := fw.FilterQuery(context.Background(), "did:key:peer-1", "echo")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrGlobalRateLimitExceeded)
	})
}

func TestWave41FilterQueryReputationBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		assess     *ReputationAssessment
		checkErr   error
		minScore   float64
		wantErr    string
		wantNoErr  bool
		wantCalled bool
	}{
		{
			name:       "reputation checker error falls through to ACL",
			checkErr:   errors.New("temporary reputation outage"),
			minScore:   0.5,
			wantNoErr:  true,
			wantCalled: true,
		},
		{
			name:       "nil assessment falls through to ACL",
			minScore:   0.5,
			wantNoErr:  true,
			wantCalled: true,
		},
		{
			name: "temporarily unsafe peer is denied",
			assess: &ReputationAssessment{
				Score:             0.9,
				TemporarilyUnsafe: true,
			},
			minScore:   0.5,
			wantErr:    "temporarily unsafe",
			wantCalled: true,
		},
		{
			name: "returning peer requiring approval is denied",
			assess: &ReputationAssessment{
				Score:            0.9,
				ReturningPeer:    true,
				Allowed:          true,
				RequiresApproval: true,
				TrustEntryState:  "pending_review",
			},
			minScore:   0.5,
			wantErr:    "requires review",
			wantCalled: true,
		},
		{
			name: "returning disallowed peer is denied",
			assess: &ReputationAssessment{
				Score:           0.9,
				ReturningPeer:   true,
				Allowed:         false,
				TrustEntryState: "blocked",
			},
			minScore:   0.5,
			wantErr:    "requires review",
			wantCalled: true,
		},
		{
			name: "low nonzero score is denied",
			assess: &ReputationAssessment{
				Score:   0.2,
				Allowed: true,
			},
			minScore:   0.5,
			wantErr:    "below minimum",
			wantCalled: true,
		},
		{
			name: "zero bootstrap score falls through to ACL",
			assess: &ReputationAssessment{
				Score:   0,
				Allowed: true,
			},
			minScore:   0.5,
			wantNoErr:  true,
			wantCalled: true,
		},
		{
			name: "healthy assessment falls through to ACL",
			assess: &ReputationAssessment{
				Score:         0.8,
				ReturningPeer: true,
				Allowed:       true,
			},
			minScore:   0.5,
			wantNoErr:  true,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fw := New([]ACLRule{
				{PeerDID: "did:key:peer-1", Action: ACLActionAllow, Tools: []string{"echo"}},
			}, testLogger())

			var called bool
			fw.SetReputationChecker(func(ctx context.Context, peerDID string) (*ReputationAssessment, error) {
				require.NotNil(t, ctx)
				assert.Equal(t, "did:key:peer-1", peerDID)
				called = true
				return tt.assess, tt.checkErr
			}, tt.minScore)

			err := fw.FilterQuery(context.Background(), "did:key:peer-1", "echo")
			assert.Equal(t, tt.wantCalled, called)
			if tt.wantNoErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestWave41SanitizeResponseStripsNestedSensitiveFieldsAndRedactsPaths(t *testing.T) {
	t.Parallel()

	fw := New(nil, testLogger())

	response := map[string]interface{}{
		"answer":       "Loaded from /Users/alice/work/lango/private.txt",
		"db_path":      "/tmp/lango.db",
		"apiToken":     "secret-token",
		"private_key":  "private-key",
		"filePath":     "/Users/alice/secret/file.txt",
		"safe_counter": 3,
		"nested": map[string]interface{}{
			"internalID": "internal-123",
			"note":       "Trace stored at /var/lib/lango/internal/run.log",
			"safe":       "visible",
		},
	}

	sanitized := fw.SanitizeResponse(response)

	assert.NotContains(t, sanitized, "db_path")
	assert.NotContains(t, sanitized, "apiToken")
	assert.NotContains(t, sanitized, "private_key")
	assert.NotContains(t, sanitized, "filePath")
	assert.Equal(t, 3, sanitized["safe_counter"])
	assert.Equal(t, "Loaded from [path-redacted]", sanitized["answer"])

	nested := sanitized["nested"].(map[string]interface{})
	assert.NotContains(t, nested, "internalID")
	assert.Equal(t, "Trace stored at [path-redacted]", nested["note"])
	assert.Equal(t, "visible", nested["safe"])
}

func TestWave41AttestResponseNilAndConfiguredFunction(t *testing.T) {
	t.Parallel()

	fw := New(nil, testLogger())
	result, err := fw.AttestResponse([]byte("response"), []byte("agent"))
	require.NoError(t, err)
	assert.Nil(t, result)

	want := &AttestationResult{
		Proof:        []byte("proof"),
		PublicInputs: []byte("public"),
		CircuitID:    "response-attestation",
		Scheme:       "groth16",
	}
	fw.SetZKAttestFunc(func(responseHash, agentDIDHash []byte) (*AttestationResult, error) {
		expectedResponseHash := sha256.Sum256([]byte("response"))
		expectedAgentDIDHash := sha256.Sum256([]byte("agent"))
		assert.True(t, bytes.Equal(expectedResponseHash[:], responseHash))
		assert.True(t, bytes.Equal(expectedAgentDIDHash[:], agentDIDHash))
		return want, nil
	})

	result, err = fw.AttestResponse([]byte("response"), []byte("agent"))
	require.NoError(t, err)
	assert.Same(t, want, result)

	wantErr := errors.New("attestation failed")
	fw.SetZKAttestFunc(func(_, _ []byte) (*AttestationResult, error) {
		return nil, wantErr
	})

	result, err = fw.AttestResponse([]byte("response"), []byte("agent"))
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, result)
}

func TestWave41MatchingHelpers(t *testing.T) {
	t.Parallel()

	assert.True(t, matchesPeer(WildcardAll, "did:key:any"))
	assert.True(t, matchesPeer("did:key:peer-1", "did:key:peer-1"))
	assert.False(t, matchesPeer("did:key:peer-1", "did:key:peer-2"))

	assert.True(t, matchesTool(nil, "echo"))
	assert.True(t, matchesTool([]string{}, "echo"))
	assert.True(t, matchesTool([]string{WildcardAll}, "echo"))
	assert.True(t, matchesTool([]string{"search*"}, "search.web"))
	assert.True(t, matchesTool([]string{"echo"}, "echo"))
	assert.False(t, matchesTool([]string{"search*"}, "echo"))
	assert.False(t, matchesTool([]string{"echo"}, "echo.extra"))
}
