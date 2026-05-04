package config

import (
	"encoding/json"
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

func TestDefaultConfig_OntologyAdmissionObserveDefaults(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, "off", cfg.Ontology.Governance.AdmissionMode)
	assert.InDelta(t, 0.60, cfg.Ontology.Governance.LearningDefaultConfidence, 0.001)
	assert.InDelta(t, 0.50, cfg.Ontology.Governance.LibrarianDefaultConfidence, 0.001)
}

func TestOntologyAdmissionConfidenceJSONRoundTripPreservesZero(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Ontology.Governance.AdmissionMode = OntologyAdmissionModeObserve
	cfg.Ontology.Governance.LearningDefaultConfidence = 0.0
	cfg.Ontology.Governance.LibrarianDefaultConfidence = 0.0

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"learningDefaultConfidence":0`)
	assert.Contains(t, string(data), `"librarianDefaultConfidence":0`)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")
	require.NoError(t, os.WriteFile(cfgPath, data, 0644))

	result, err := Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result.Config.Ontology.Governance.LearningDefaultConfidence)
	assert.Equal(t, 0.0, result.Config.Ontology.Governance.LibrarianDefaultConfidence)
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

	t.Run("backfills legacy ontology admission defaults after direct unmarshal", func(t *testing.T) {
		t.Parallel()

		var cfg Config
		content := `{
			"server": { "port": 18789 },
			"logging": { "level": "info", "format": "console" },
			"ontology": {
				"governance": {}
			}
		}`
		require.NoError(t, json.Unmarshal([]byte(content), &cfg))

		require.NoError(t, PostLoad(&cfg))
		assert.Equal(t, OntologyAdmissionModeOff, cfg.Ontology.Governance.AdmissionMode)
		assert.InDelta(t, OntologyLearningDefaultConfidenceFallback, cfg.Ontology.Governance.LearningDefaultConfidence, 0.001)
		assert.InDelta(t, OntologyLibrarianDefaultConfidenceFallback, cfg.Ontology.Governance.LibrarianDefaultConfidence, 0.001)
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

	// Validate rejects sandbox.workspacePath nested under cfg.DataRoot.
	// This is the defense against the regression where a relative workspace
	// path is normalized under DataRoot and then collides with the
	// DefaultToolPolicy control-plane deny, making the workspace unreachable.
	t.Run("sandbox workspacePath under DataRoot rejected", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.DataRoot = "/tmp/lango-test-dataroot"
		cfg.Sandbox.WorkspacePath = "/tmp/lango-test-dataroot/repo"
		err := Validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox.workspacePath")
		assert.Contains(t, err.Error(), "inside cfg.DataRoot")
	})

	t.Run("sandbox workspacePath equal to DataRoot rejected", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.DataRoot = "/tmp/lango-test-dataroot"
		cfg.Sandbox.WorkspacePath = "/tmp/lango-test-dataroot"
		err := Validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox.workspacePath")
	})

	t.Run("sandbox workspacePath outside DataRoot accepted", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.DataRoot = "/tmp/lango-test-dataroot"
		cfg.Sandbox.WorkspacePath = "/tmp/some-other-dir"
		assert.NoError(t, Validate(cfg))
	})

	t.Run("sandbox allowedWritePaths entry under DataRoot rejected", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.DataRoot = "/tmp/lango-test-dataroot"
		cfg.Sandbox.AllowedWritePaths = []string{"/tmp/outside", "/tmp/lango-test-dataroot/scratch"}
		err := Validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox.allowedWritePaths")
		assert.Contains(t, err.Error(), "scratch")
	})

	t.Run("sandbox workspacePath empty accepted", func(t *testing.T) {
		t.Parallel()

		// Empty WorkspacePath is valid — supervisor falls back to os.Getwd().
		cfg := DefaultConfig()
		cfg.DataRoot = "/tmp/lango-test-dataroot"
		cfg.Sandbox.WorkspacePath = ""
		assert.NoError(t, Validate(cfg))
	})

	t.Run("ontology admission mode rejects unknown value", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Ontology.Governance.AdmissionMode = "enforce"

		err := Validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ontology.governance.admissionMode")
		assert.Contains(t, err.Error(), "off, observe")
	})

	t.Run("ontology admission mode empty accepted for backward compatibility", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Ontology.Governance.AdmissionMode = ""

		assert.NoError(t, Validate(cfg))
	})

	t.Run("ontology learning confidence rejects out of range", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Ontology.Governance.AdmissionMode = OntologyAdmissionModeObserve
		cfg.Ontology.Governance.LearningDefaultConfidence = 1.25

		err := Validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ontology.governance.learningDefaultConfidence")
		assert.Contains(t, err.Error(), "0.0-1.0")
	})

	t.Run("ontology librarian confidence rejects out of range", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Ontology.Governance.AdmissionMode = OntologyAdmissionModeObserve
		cfg.Ontology.Governance.LibrarianDefaultConfidence = -0.10

		err := Validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ontology.governance.librarianDefaultConfidence")
		assert.Contains(t, err.Error(), "0.0-1.0")
	})

	t.Run("ontology admission off allows legacy confidence values", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Ontology.Governance.AdmissionMode = OntologyAdmissionModeOff
		cfg.Ontology.Governance.LearningDefaultConfidence = 1.25
		cfg.Ontology.Governance.LibrarianDefaultConfidence = -0.10

		assert.NoError(t, Validate(cfg))
	})
}

func TestPathIsUnder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveChild  string
		giveParent string
		want       bool
	}{
		{give: "nested", giveChild: "/a/b/c", giveParent: "/a/b", want: true},
		{give: "same path", giveChild: "/a/b", giveParent: "/a/b", want: true},
		{give: "outside", giveChild: "/a/c", giveParent: "/a/b", want: false},
		{give: "parent is child", giveChild: "/a", giveParent: "/a/b", want: false},
		{give: "sibling", giveChild: "/other", giveParent: "/a/b", want: false},
		{give: "empty child", giveChild: "", giveParent: "/a", want: false},
		{give: "empty parent", giveChild: "/a", giveParent: "", want: false},
		{give: "nested with trailing separator on child", giveChild: "/a/b/c/", giveParent: "/a/b", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := pathIsUnder(tt.giveChild, tt.giveParent)
			assert.Equal(t, tt.want, got)
		})
	}
}
