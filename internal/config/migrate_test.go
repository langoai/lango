package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig_ApprovalPolicy(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, ApprovalPolicyDangerous, cfg.Security.Interceptor.ApprovalPolicy)
	assert.True(t, cfg.Security.Interceptor.Enabled)
}
