package firewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

func TestScanAndRedact_ExactTerms(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{
		OwnerName:  "Alice Kim",
		OwnerEmail: "alice@example.com",
		OwnerPhone: "010-1234-5678",
	}, testLogger())

	tests := []struct {
		give     string
		giveData map[string]interface{}
		wantKeys []string
	}{
		{
			give: "name redacted",
			giveData: map[string]interface{}{
				"result": "Contact Alice Kim for details",
			},
			wantKeys: []string{"result"},
		},
		{
			give: "email redacted",
			giveData: map[string]interface{}{
				"contact": "Send mail to alice@example.com",
			},
			wantKeys: []string{"contact"},
		},
		{
			give: "phone redacted",
			giveData: map[string]interface{}{
				"phone": "Call 010-1234-5678",
			},
			wantKeys: []string{"phone"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, blocked := shield.ScanAndRedact(tt.giveData)
			require.Len(t, blocked, len(tt.wantKeys))
			for _, key := range tt.wantKeys {
				assert.Equal(t, redactedPlaceholder, result[key])
				assert.Contains(t, blocked, key)
			}
		})
	}
}

func TestScanAndRedact_RegexPatterns(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{}, testLogger())

	tests := []struct {
		give      string
		giveData  map[string]interface{}
		wantBlock bool
	}{
		{
			give:      "generic email caught",
			giveData:  map[string]interface{}{"info": "Email bob@corp.io for help"},
			wantBlock: true,
		},
		{
			give:      "generic phone caught",
			giveData:  map[string]interface{}{"info": "Call 02-555-1234 now"},
			wantBlock: true,
		},
		{
			give:      "phone with dots caught",
			giveData:  map[string]interface{}{"info": "Phone: 010.9876.5432"},
			wantBlock: true,
		},
		{
			give:      "no match passes through",
			giveData:  map[string]interface{}{"info": "Nothing sensitive here"},
			wantBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			_, blocked := shield.ScanAndRedact(tt.giveData)
			if tt.wantBlock {
				assert.NotEmpty(t, blocked)
			} else {
				assert.Empty(t, blocked)
			}
		})
	}
}

func TestScanAndRedact_ConversationBlocking(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{
		BlockConversations: true,
	}, testLogger())

	tests := []struct {
		give     string
		giveData map[string]interface{}
		wantKey  string
	}{
		{
			give:     "conversation key blocked",
			giveData: map[string]interface{}{"conversation": "secret chat content"},
			wantKey:  "conversation",
		},
		{
			give:     "message_history key blocked",
			giveData: map[string]interface{}{"message_history": []interface{}{"msg1", "msg2"}},
			wantKey:  "message_history",
		},
		{
			give:     "chat_log key blocked",
			giveData: map[string]interface{}{"chat_log": "some log data"},
			wantKey:  "chat_log",
		},
		{
			give:     "session_history key blocked",
			giveData: map[string]interface{}{"session_history": "session data"},
			wantKey:  "session_history",
		},
		{
			give:     "chat_history key blocked",
			giveData: map[string]interface{}{"chat_history": map[string]interface{}{"key": "val"}},
			wantKey:  "chat_history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, blocked := shield.ScanAndRedact(tt.giveData)
			require.Len(t, blocked, 1)
			assert.Equal(t, tt.wantKey, blocked[0])
			assert.Equal(t, redactedPlaceholder, result[tt.wantKey])
		})
	}
}

func TestScanAndRedact_ConversationBlocking_Disabled(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{
		BlockConversations: false,
	}, testLogger())

	data := map[string]interface{}{
		"conversation": "safe content without PII",
	}

	result, blocked := shield.ScanAndRedact(data)
	assert.Empty(t, blocked)
	assert.Equal(t, "safe content without PII", result["conversation"])
}

func TestScanAndRedact_NestedMaps(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{
		OwnerName: "Alice Kim",
	}, testLogger())

	data := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "Alice Kim is the owner",
			"safe":  "nothing here",
		},
		"list": []interface{}{
			"Alice Kim was mentioned",
			"clean item",
			map[string]interface{}{
				"deep": "deep mention of Alice Kim",
			},
		},
	}

	result, blocked := shield.ScanAndRedact(data)
	require.Len(t, blocked, 3)
	assert.Contains(t, blocked, "outer.inner")
	assert.Contains(t, blocked, "list[0]")
	assert.Contains(t, blocked, "list[2].deep")

	outer := result["outer"].(map[string]interface{})
	assert.Equal(t, redactedPlaceholder, outer["inner"])
	assert.Equal(t, "nothing here", outer["safe"])

	list := result["list"].([]interface{})
	assert.Equal(t, redactedPlaceholder, list[0])
	assert.Equal(t, "clean item", list[1])
	deepMap := list[2].(map[string]interface{})
	assert.Equal(t, redactedPlaceholder, deepMap["deep"])
}

func TestScanAndRedact_NoMatch(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{
		OwnerName:          "Alice Kim",
		OwnerEmail:         "alice@example.com",
		BlockConversations: true,
	}, testLogger())

	data := map[string]interface{}{
		"result":  "The weather is sunny today",
		"count":   42,
		"details": map[string]interface{}{"note": "no PII here"},
	}

	result, blocked := shield.ScanAndRedact(data)
	assert.Empty(t, blocked)
	assert.Equal(t, "The weather is sunny today", result["result"])
	assert.Equal(t, 42, result["count"])
	details := result["details"].(map[string]interface{})
	assert.Equal(t, "no PII here", details["note"])
}

func TestContainsOwnerData(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{
		OwnerName:  "Alice Kim",
		OwnerEmail: "alice@example.com",
		OwnerPhone: "010-1234-5678",
		ExtraTerms: []string{"Project Omega"},
	}, testLogger())

	tests := []struct {
		give string
		want bool
	}{
		{give: "Contact Alice Kim please", want: true},
		{give: "Send to alice@example.com", want: true},
		{give: "Call 010-1234-5678", want: true},
		{give: "Top secret Project Omega data", want: true},
		{give: "case insensitive ALICE KIM test", want: true},
		{give: "generic email test@domain.org", want: true},
		{give: "generic phone 02-555-1234", want: true},
		{give: "nothing sensitive at all", want: false},
		{give: "just a plain number 42", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, shield.ContainsOwnerData(tt.give))
		})
	}
}

func TestNewOwnerShield_EmptyConfig(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{}, testLogger())

	assert.Empty(t, shield.exactTerms)
	assert.Len(t, shield.regexPatterns, 2)
	assert.False(t, shield.blockConvKeys)
}

func TestNewOwnerShield_ExtraTerms(t *testing.T) {
	shield := NewOwnerShield(OwnerProtectionConfig{
		ExtraTerms: []string{"secret-project", "", "codename-alpha"},
	}, testLogger())

	assert.Len(t, shield.exactTerms, 2)
	assert.Contains(t, shield.exactTerms, "secret-project")
	assert.Contains(t, shield.exactTerms, "codename-alpha")
}
