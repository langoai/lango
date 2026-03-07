package firewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestValidateRule_AllowWildcardPeerAndTools(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    ACLRule
		wantErr bool
	}{
		{
			give:    ACLRule{PeerDID: WildcardAll, Action: ACLActionAllow},
			wantErr: true, // wildcard peer + empty tools (= all)
		},
		{
			give:    ACLRule{PeerDID: WildcardAll, Action: ACLActionAllow, Tools: []string{WildcardAll}},
			wantErr: true, // wildcard peer + wildcard tool
		},
		{
			give:    ACLRule{PeerDID: WildcardAll, Action: ACLActionAllow, Tools: []string{"echo", WildcardAll}},
			wantErr: true, // wildcard tool mixed in
		},
		{
			give:    ACLRule{PeerDID: WildcardAll, Action: ACLActionDeny},
			wantErr: false, // deny rules always safe
		},
		{
			give:    ACLRule{PeerDID: WildcardAll, Action: ACLActionDeny, Tools: []string{WildcardAll}},
			wantErr: false, // deny rules always safe
		},
		{
			give:    ACLRule{PeerDID: "did:key:specific", Action: ACLActionAllow, Tools: []string{WildcardAll}},
			wantErr: false, // specific peer OK
		},
		{
			give:    ACLRule{PeerDID: WildcardAll, Action: ACLActionAllow, Tools: []string{"echo"}},
			wantErr: false, // specific tool OK
		},
		{
			give:    ACLRule{PeerDID: "did:key:abc", Action: ACLActionAllow},
			wantErr: false, // specific peer, all tools
		},
	}

	for _, tt := range tests {
		t.Run(tt.give.PeerDID+"/"+string(tt.give.Action), func(t *testing.T) {
			t.Parallel()
			err := ValidateRule(tt.give)
			if tt.wantErr {
				assert.Error(t, err, "expected error for overly permissive rule")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddRule_RejectsOverlyPermissive(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	fw := New(nil, logger.Sugar())

	err := fw.AddRule(ACLRule{PeerDID: WildcardAll, Action: ACLActionAllow, Tools: []string{WildcardAll}})
	assert.Error(t, err, "expected AddRule to reject wildcard allow rule")

	// Verify the rule was NOT added.
	rules := fw.Rules()
	assert.Empty(t, rules)
}

func TestAddRule_AcceptsValidRule(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	fw := New(nil, logger.Sugar())

	err := fw.AddRule(ACLRule{PeerDID: "did:key:peer-1", Action: ACLActionAllow, Tools: []string{"echo"}})
	require.NoError(t, err)

	rules := fw.Rules()
	require.Len(t, rules, 1)
	assert.Equal(t, "did:key:peer-1", rules[0].PeerDID)
}

func TestAddRule_AcceptsDenyWildcard(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	fw := New(nil, logger.Sugar())

	err := fw.AddRule(ACLRule{PeerDID: WildcardAll, Action: ACLActionDeny, Tools: []string{WildcardAll}})
	require.NoError(t, err)

	rules := fw.Rules()
	require.Len(t, rules, 1)
}

func TestNew_WarnsOnOverlyPermissiveInitialRules(t *testing.T) {
	t.Parallel()

	// Should not panic — just logs a warning for backward compatibility.
	logger, _ := zap.NewDevelopment()
	fw := New([]ACLRule{
		{PeerDID: WildcardAll, Action: ACLActionAllow},
	}, logger.Sugar())

	// Rule is still loaded (backward compat).
	rules := fw.Rules()
	require.Len(t, rules, 1)
}
