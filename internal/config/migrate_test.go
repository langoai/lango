package config

import (
	"testing"
)

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
