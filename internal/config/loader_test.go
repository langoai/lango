package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, 18789, cfg.Server.Port)
	assert.Equal(t, "anthropic", cfg.Agent.Provider)
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestExpandEnvVars(t *testing.T) {
	t.Parallel()

	t.Run("expands existing env var", func(t *testing.T) {
		t.Parallel()

		os.Setenv("TEST_API_KEY_EXPAND", "sk-test-123")
		defer os.Unsetenv("TEST_API_KEY_EXPAND")

		result := ExpandEnvVars("${TEST_API_KEY_EXPAND}")
		assert.Equal(t, "sk-test-123", result)
	})

	t.Run("keeps non-existent var unchanged", func(t *testing.T) {
		t.Parallel()

		result := ExpandEnvVars("${NON_EXISTENT_VAR}")
		assert.Equal(t, "${NON_EXISTENT_VAR}", result)
	})
}

func TestPostLoad(t *testing.T) {
	t.Parallel()

	t.Run("applies full processing chain", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		err := PostLoad(cfg)
		require.NoError(t, err)

		// Paths should be normalized to absolute.
		assert.True(t, filepath.IsAbs(cfg.DataRoot), "DataRoot should be absolute")
		assert.True(t, filepath.IsAbs(cfg.Session.DatabasePath), "Session.DatabasePath should be absolute")
		assert.True(t, filepath.IsAbs(cfg.Skill.SkillsDir), "Skill.SkillsDir should be absolute")
	})

	t.Run("idempotent", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		err := PostLoad(cfg)
		require.NoError(t, err)

		// Snapshot after first call.
		dataRoot1 := cfg.DataRoot
		dbPath1 := cfg.Session.DatabasePath
		skillsDir1 := cfg.Skill.SkillsDir

		// Second call should produce identical result.
		err = PostLoad(cfg)
		require.NoError(t, err)

		assert.Equal(t, dataRoot1, cfg.DataRoot)
		assert.Equal(t, dbPath1, cfg.Session.DatabasePath)
		assert.Equal(t, skillsDir1, cfg.Skill.SkillsDir)
	})

	t.Run("rejects invalid config", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Payment.Enabled = true
		cfg.Payment.Network.RPCURL = "" // missing required field
		err := PostLoad(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "payment.network.rpcUrl")
	})
}

func TestNormalizePaths_Sandbox(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	t.Run("WorkspacePath tilde expanded", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Sandbox.WorkspacePath = "~/work"
		NormalizePaths(cfg)

		assert.Equal(t, filepath.Join(home, "work"), cfg.Sandbox.WorkspacePath)
	})

	t.Run("WorkspacePath relative resolved under DataRoot", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Sandbox.WorkspacePath = "workspace"
		NormalizePaths(cfg)

		assert.Equal(t, filepath.Join(cfg.DataRoot, "workspace"), cfg.Sandbox.WorkspacePath)
	})

	t.Run("WorkspacePath empty stays empty", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Sandbox.WorkspacePath = ""
		NormalizePaths(cfg)

		assert.Empty(t, cfg.Sandbox.WorkspacePath)
	})

	t.Run("WorkspacePath absolute unchanged", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Sandbox.WorkspacePath = "/abs/work"
		NormalizePaths(cfg)

		assert.Equal(t, "/abs/work", cfg.Sandbox.WorkspacePath)
	})

	t.Run("AllowedWritePaths slice normalized in place", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Sandbox.AllowedWritePaths = []string{"~/a", "/b", "rel"}
		NormalizePaths(cfg)

		require.Len(t, cfg.Sandbox.AllowedWritePaths, 3)
		assert.Equal(t, filepath.Join(home, "a"), cfg.Sandbox.AllowedWritePaths[0])
		assert.Equal(t, "/b", cfg.Sandbox.AllowedWritePaths[1])
		assert.Equal(t, filepath.Join(cfg.DataRoot, "rel"), cfg.Sandbox.AllowedWritePaths[2])
	})

	t.Run("AllowedWritePaths empty slice unchanged", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Sandbox.AllowedWritePaths = nil
		NormalizePaths(cfg)

		assert.Nil(t, cfg.Sandbox.AllowedWritePaths)
	})

	t.Run("SeatbeltCustomProfile tilde expanded", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Sandbox.OS.SeatbeltCustomProfile = "~/profile.sb"
		NormalizePaths(cfg)

		assert.Equal(t, filepath.Join(home, "profile.sb"), cfg.Sandbox.OS.SeatbeltCustomProfile)
	})

	t.Run("idempotent on sandbox paths", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Sandbox.WorkspacePath = "~/work"
		cfg.Sandbox.AllowedWritePaths = []string{"~/a", "rel"}
		cfg.Sandbox.OS.SeatbeltCustomProfile = "~/profile.sb"

		NormalizePaths(cfg)
		first := struct {
			ws       string
			allowed  []string
			seatbelt string
		}{
			cfg.Sandbox.WorkspacePath,
			append([]string(nil), cfg.Sandbox.AllowedWritePaths...),
			cfg.Sandbox.OS.SeatbeltCustomProfile,
		}

		NormalizePaths(cfg)
		assert.Equal(t, first.ws, cfg.Sandbox.WorkspacePath)
		assert.Equal(t, first.allowed, cfg.Sandbox.AllowedWritePaths)
		assert.Equal(t, first.seatbelt, cfg.Sandbox.OS.SeatbeltCustomProfile)
	})
}

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		require.NoError(t, Validate(cfg))
	})

	t.Run("invalid port", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Server.Port = 0
		assert.Error(t, Validate(cfg))
	})

	t.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Agent.Provider = "invalid"
		cfg.Providers = map[string]ProviderConfig{
			"google": {Type: "gemini", APIKey: "test"},
		}
		assert.Error(t, Validate(cfg))
	})

	t.Run("invalid log level", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Logging.Level = "invalid"
		assert.Error(t, Validate(cfg))
	})
}
