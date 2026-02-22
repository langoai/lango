package app

import (
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
)

func TestNeedsApproval(t *testing.T) {
	tests := []struct {
		give string
		tool *agent.Tool
		ic   config.InterceptorConfig
		want bool
	}{
		{
			give: "dangerous policy + dangerous tool → true",
			tool: &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelDangerous},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous},
			want: true,
		},
		{
			give: "dangerous policy + safe tool → false",
			tool: &agent.Tool{Name: "fs_read", SafetyLevel: agent.SafetyLevelSafe},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous},
			want: false,
		},
		{
			give: "dangerous policy + moderate tool → false",
			tool: &agent.Tool{Name: "fs_mkdir", SafetyLevel: agent.SafetyLevelModerate},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous},
			want: false,
		},
		{
			give: "dangerous policy + zero value SafetyLevel → true (fail-safe)",
			tool: &agent.Tool{Name: "unknown"},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous},
			want: true,
		},
		{
			give: "all policy + safe tool → true",
			tool: &agent.Tool{Name: "fs_read", SafetyLevel: agent.SafetyLevelSafe},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll},
			want: true,
		},
		{
			give: "configured policy + unlisted tool → false",
			tool: &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelDangerous},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyConfigured, SensitiveTools: []string{"fs_delete"}},
			want: false,
		},
		{
			give: "configured policy + listed tool → true",
			tool: &agent.Tool{Name: "fs_delete", SafetyLevel: agent.SafetyLevelDangerous},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyConfigured, SensitiveTools: []string{"fs_delete"}},
			want: true,
		},
		{
			give: "none policy + dangerous tool → false",
			tool: &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelDangerous},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyNone},
			want: false,
		},
		{
			give: "exempt tool overrides dangerous policy",
			tool: &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelDangerous},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous, ExemptTools: []string{"exec"}},
			want: false,
		},
		{
			give: "exempt tool overrides all policy",
			tool: &agent.Tool{Name: "fs_read", SafetyLevel: agent.SafetyLevelSafe},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll, ExemptTools: []string{"fs_read"}},
			want: false,
		},
		{
			give: "sensitive tool overrides configured-none",
			tool: &agent.Tool{Name: "fs_read", SafetyLevel: agent.SafetyLevelSafe},
			ic:   config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous, SensitiveTools: []string{"fs_read"}},
			want: true,
		},
		{
			give: "exempt takes priority over sensitive",
			tool: &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelDangerous},
			ic: config.InterceptorConfig{
				ApprovalPolicy: config.ApprovalPolicyAll,
				SensitiveTools: []string{"exec"},
				ExemptTools:    []string{"exec"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := needsApproval(tt.tool, tt.ic)
			if got != tt.want {
				t.Errorf("needsApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildApprovalSummary(t *testing.T) {
	tests := []struct {
		give     string
		toolName string
		params   map[string]interface{}
		want     string
	}{
		{
			give:     "exec command",
			toolName: "exec",
			params:   map[string]interface{}{"command": "ls -la"},
			want:     "Execute: ls -la",
		},
		{
			give:     "exec_bg command",
			toolName: "exec_bg",
			params:   map[string]interface{}{"command": "sleep 10"},
			want:     "Execute: sleep 10",
		},
		{
			give:     "fs_write",
			toolName: "fs_write",
			params:   map[string]interface{}{"path": "/tmp/test.txt", "content": "hello"},
			want:     "Write to /tmp/test.txt (5 bytes)",
		},
		{
			give:     "fs_edit",
			toolName: "fs_edit",
			params:   map[string]interface{}{"path": "/tmp/test.txt"},
			want:     "Edit file: /tmp/test.txt",
		},
		{
			give:     "fs_delete",
			toolName: "fs_delete",
			params:   map[string]interface{}{"path": "/tmp/test.txt"},
			want:     "Delete: /tmp/test.txt",
		},
		{
			give:     "browser_navigate",
			toolName: "browser_navigate",
			params:   map[string]interface{}{"url": "https://example.com"},
			want:     "Navigate to: https://example.com",
		},
		{
			give:     "browser_action with selector",
			toolName: "browser_action",
			params:   map[string]interface{}{"action": "click", "selector": "#submit"},
			want:     "Browser click on: #submit",
		},
		{
			give:     "browser_action without selector",
			toolName: "browser_action",
			params:   map[string]interface{}{"action": "eval"},
			want:     "Browser action: eval",
		},
		{
			give:     "secrets_store",
			toolName: "secrets_store",
			params:   map[string]interface{}{"name": "api_key"},
			want:     "Store secret: api_key",
		},
		{
			give:     "secrets_get",
			toolName: "secrets_get",
			params:   map[string]interface{}{"name": "api_key"},
			want:     "Retrieve secret: api_key",
		},
		{
			give:     "secrets_delete",
			toolName: "secrets_delete",
			params:   map[string]interface{}{"name": "api_key"},
			want:     "Delete secret: api_key",
		},
		{
			give:     "crypto_encrypt",
			toolName: "crypto_encrypt",
			params:   map[string]interface{}{"data": "secret"},
			want:     "Encrypt data",
		},
		{
			give:     "crypto_decrypt",
			toolName: "crypto_decrypt",
			params:   map[string]interface{}{"ciphertext": "abc"},
			want:     "Decrypt ciphertext",
		},
		{
			give:     "crypto_sign",
			toolName: "crypto_sign",
			params:   map[string]interface{}{"data": "msg"},
			want:     "Generate digital signature",
		},
		{
			give:     "unknown tool",
			toolName: "custom_tool",
			params:   map[string]interface{}{},
			want:     "Tool: custom_tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := buildApprovalSummary(tt.toolName, tt.params)
			if got != tt.want {
				t.Errorf("buildApprovalSummary(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		give   string
		maxLen int
		want   string
	}{
		{give: "short", maxLen: 10, want: "short"},
		{give: "exactly10!", maxLen: 10, want: "exactly10!"},
		{give: "this is longer than ten", maxLen: 10, want: "this is lo..."},
		{give: "", maxLen: 5, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := truncate(tt.give, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.give, tt.maxLen, got, tt.want)
			}
		})
	}
}
