package config

import (
	"testing"
)

func TestMigrateApprovalPolicy(t *testing.T) {
	tests := []struct {
		give string
		ic   InterceptorConfig
		want ApprovalPolicy
	}{
		{
			give: "already set → keep as is",
			ic:   InterceptorConfig{ApprovalPolicy: ApprovalPolicyAll},
			want: ApprovalPolicyAll,
		},
		{
			give: "approvalRequired=true + sensitiveTools → configured",
			ic: InterceptorConfig{
				ApprovalRequired: true,
				SensitiveTools:   []string{"exec"},
			},
			want: ApprovalPolicyConfigured,
		},
		{
			give: "approvalRequired=true + no sensitiveTools → dangerous",
			ic:   InterceptorConfig{ApprovalRequired: true},
			want: ApprovalPolicyDangerous,
		},
		{
			give: "approvalRequired=false → empty (inherits default from viper)",
			ic:   InterceptorConfig{ApprovalRequired: false},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cfg := &Config{
				Security: SecurityConfig{
					Interceptor: tt.ic,
				},
			}
			migrateApprovalPolicy(cfg)
			got := cfg.Security.Interceptor.ApprovalPolicy
			if got != tt.want {
				t.Errorf("migrateApprovalPolicy() → %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_ApprovalPolicy(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Security.Interceptor.ApprovalPolicy != ApprovalPolicyDangerous {
		t.Errorf("expected default approval policy %q, got %q",
			ApprovalPolicyDangerous, cfg.Security.Interceptor.ApprovalPolicy)
	}

	if !cfg.Security.Interceptor.Enabled {
		t.Error("expected default interceptor enabled to be true")
	}
}
