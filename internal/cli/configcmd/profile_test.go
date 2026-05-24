package configcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/storage"
)

type mutableProfileStore struct {
	profiles map[string]*config.Config
	active   string
	infos    []configstore.ProfileInfo
	explicit map[string]map[string]bool

	saveErr      error
	loadErr      error
	setActiveErr error
	listErr      error
	deleteErr    error
	existsErr    error
}

func (m *mutableProfileStore) Save(_ context.Context, name string, cfg *config.Config, explicitKeys map[string]bool) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.profiles == nil {
		m.profiles = make(map[string]*config.Config)
	}
	if m.explicit == nil {
		m.explicit = make(map[string]map[string]bool)
	}
	m.profiles[name] = cfg
	m.explicit[name] = explicitKeys
	if m.active == "" {
		m.active = name
	}
	return nil
}

func (m *mutableProfileStore) Load(_ context.Context, name string) (*config.Config, map[string]bool, error) {
	if m.loadErr != nil {
		return nil, nil, m.loadErr
	}
	cfg, ok := m.profiles[name]
	if !ok {
		return nil, nil, errors.New("not found")
	}
	return cfg, nil, nil
}

func (m *mutableProfileStore) LoadActive(_ context.Context) (string, *config.Config, map[string]bool, error) {
	cfg, ok := m.profiles[m.active]
	if !ok {
		return "", nil, nil, errors.New("not found")
	}
	return m.active, cfg, nil, nil
}

func (m *mutableProfileStore) SetActive(_ context.Context, name string) error {
	if m.setActiveErr != nil {
		return m.setActiveErr
	}
	if _, ok := m.profiles[name]; !ok {
		return errors.New("not found")
	}
	m.active = name
	return nil
}

func (m *mutableProfileStore) List(_ context.Context) ([]configstore.ProfileInfo, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.infos, nil
}

func (m *mutableProfileStore) Delete(_ context.Context, name string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.profiles, name)
	return nil
}

func (m *mutableProfileStore) Exists(_ context.Context, name string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	_, ok := m.profiles[name]
	return ok, nil
}

func executeConfigProfileCommand(t *testing.T, cmd *cobra.Command, input string, args ...string) (string, string, error) {
	t.Helper()
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if input != "" {
		cmd.SetIn(bytes.NewBufferString(input))
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func TestConfigProfileCommands_WriteToCommandStreams(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	store := &mutableProfileStore{
		profiles: map[string]*config.Config{
			"default": config.DefaultConfig(),
		},
		active: "default",
		infos: []configstore.ProfileInfo{{
			Name:      "default",
			Active:    true,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		}},
	}

	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      config.DefaultConfig(),
			ProfileName: store.active,
			Storage:     storage.NewFacade(store, nil),
		}, nil
	}

	{
		cmd := newListCmd(bootLoader)
		stdout, _, err := executeConfigProfileCommand(t, cmd, "")
		require.NoError(t, err)
		assert.Contains(t, stdout, "NAME")
		assert.Contains(t, stdout, "default")
	}

	{
		cmd := newCreateCmd(bootLoader)
		stdout, _, err := executeConfigProfileCommand(t, cmd, "", "staging")
		require.NoError(t, err)
		assert.Contains(t, stdout, `Profile "staging" created with default configuration.`)
	}

	{
		cmd := newUseCmd(bootLoader)
		stdout, _, err := executeConfigProfileCommand(t, cmd, "", "default")
		require.NoError(t, err)
		assert.Contains(t, stdout, `Switched to profile "default".`)
	}

	{
		filePath := filepath.Join(t.TempDir(), "config.json")
		require.NoError(t, os.WriteFile(filePath, []byte(`{"agent":{"provider":"openai"}}`), 0o600))
		cmd := newImportCmd(bootLoader)
		stdout, _, err := executeConfigProfileCommand(t, cmd, "", filePath)
		require.NoError(t, err)
		assert.Contains(t, stdout, `Imported "`)
		assert.Contains(t, stdout, `Source file deleted for security.`)
	}

	{
		cmd := newDeleteCmd(bootLoader)
		stdout, _, err := executeConfigProfileCommand(t, cmd, "n\n", "default")
		require.NoError(t, err)
		assert.Contains(t, stdout, `Delete profile "default"?`)
		assert.Contains(t, stdout, "Aborted.")
		_, ok := store.profiles["default"]
		assert.True(t, ok)
	}
}

func TestConfigDelete_ConfirmUsesCommandStreams(t *testing.T) {
	store := &mutableProfileStore{
		profiles: map[string]*config.Config{
			"staging": config.DefaultConfig(),
		},
		active: "staging",
	}

	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      config.DefaultConfig(),
			ProfileName: store.active,
			Storage:     storage.NewFacade(store, nil),
		}, nil
	}

	cmd := newDeleteCmd(bootLoader)
	stdout, stderr, err := executeConfigProfileCommand(t, cmd, "y\n", "staging")
	require.NoError(t, err)
	assert.Equal(t, "", stderr)
	assert.Contains(t, stdout, `Delete profile "staging"? This cannot be undone. [y/N]: `)
	assert.Contains(t, stdout, `Profile "staging" deleted.`)
	_, ok := store.profiles["staging"]
	assert.False(t, ok)
}

func TestConfigDelete_ForceSkipsPrompt(t *testing.T) {
	store := &mutableProfileStore{
		profiles: map[string]*config.Config{
			"staging": config.DefaultConfig(),
		},
		active: "staging",
	}

	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      config.DefaultConfig(),
			ProfileName: store.active,
			Storage:     storage.NewFacade(store, nil),
		}, nil
	}

	cmd := newDeleteCmd(bootLoader)
	stdout, stderr, err := executeConfigProfileCommand(t, cmd, "", "staging", "--force")
	require.NoError(t, err)
	assert.Equal(t, "", stderr)
	assert.NotContains(t, stdout, `Delete profile "staging"?`)
	assert.Contains(t, stdout, `Profile "staging" deleted.`)
	_, ok := store.profiles["staging"]
	assert.False(t, ok)
}

func TestConfigDelete_EOFDeniesWithoutDeleting(t *testing.T) {
	store := &mutableProfileStore{
		profiles: map[string]*config.Config{
			"staging": config.DefaultConfig(),
		},
		active: "staging",
	}

	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      config.DefaultConfig(),
			ProfileName: store.active,
			Storage:     storage.NewFacade(store, nil),
		}, nil
	}

	cmd := newDeleteCmd(bootLoader)
	stdout, stderr, err := executeConfigProfileCommand(t, cmd, "", "staging")
	require.NoError(t, err)
	assert.Equal(t, "", stderr)
	assert.Contains(t, stdout, `Delete profile "staging"? This cannot be undone. [y/N]: `)
	assert.Contains(t, stdout, "Aborted.")
	_, ok := store.profiles["staging"]
	assert.True(t, ok)
}

func TestConfigExport_WritesToConfiguredStreams(t *testing.T) {
	cfg := config.DefaultConfig()
	store := &mutableProfileStore{
		profiles: map[string]*config.Config{"default": cfg},
		active:   "default",
	}
	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			ProfileName: "default",
			Storage:     storage.NewFacade(store, nil),
		}, nil
	}
	cmd := newExportCmd(bootLoader)

	stdout, stderr, err := executeConfigProfileCommand(t, cmd, "", "default")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"agent"`)
	assert.Contains(t, stderr, "WARNING: exported configuration contains sensitive values in plaintext.")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &decoded))
	_, ok := decoded["agent"]
	assert.True(t, ok)
}

func TestNewConfigCmdWiresProfileSubcommandsAndListEmptyOutput(t *testing.T) {
	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			ProfileName: "default",
			Storage:     storage.NewFacade(&mutableProfileStore{}, nil),
		}, nil
	}
	cmd := NewConfigCmd(bootLoader)

	var names []string
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	assert.ElementsMatch(t, []string{"list", "create", "use", "delete", "import", "export", "validate"}, names)

	stdout, stderr, err := executeConfigProfileCommand(t, cmd, "", "list")
	require.NoError(t, err)
	assert.Equal(t, "", stderr)
	assert.Contains(t, stdout, "No profiles found.")
}

func TestConfigProfileCommands_CoverErrorBranches(t *testing.T) {
	bootErr := errors.New("boot failed")
	bootErrCases := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "list", cmd: newListCmd(func() (*bootstrap.Result, error) { return nil, bootErr })},
		{name: "create", cmd: newCreateCmd(func() (*bootstrap.Result, error) { return nil, bootErr }), args: []string{"new"}},
		{name: "use", cmd: newUseCmd(func() (*bootstrap.Result, error) { return nil, bootErr }), args: []string{"default"}},
		{name: "delete", cmd: newDeleteCmd(func() (*bootstrap.Result, error) { return nil, bootErr }), args: []string{"default", "--force"}},
		{name: "import", cmd: newImportCmd(func() (*bootstrap.Result, error) { return nil, bootErr }), args: []string{"config.json"}},
		{name: "export", cmd: newExportCmd(func() (*bootstrap.Result, error) { return nil, bootErr }), args: []string{"default"}},
		{name: "validate", cmd: newValidateCmd(func() (*bootstrap.Result, error) { return nil, bootErr })},
	}
	for _, tc := range bootErrCases {
		t.Run(tc.name+" boot error", func(t *testing.T) {
			_, _, err := executeConfigProfileCommand(t, tc.cmd, "", tc.args...)
			require.Error(t, err)
			assert.ErrorContains(t, err, "bootstrap: boot failed")
		})
	}

	nilStoreBoot := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Storage: storage.NewFacade(nil, nil)}, nil
	}
	nilStoreCases := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "list", cmd: newListCmd(nilStoreBoot)},
		{name: "create", cmd: newCreateCmd(nilStoreBoot), args: []string{"new"}},
		{name: "use", cmd: newUseCmd(nilStoreBoot), args: []string{"default"}},
		{name: "delete", cmd: newDeleteCmd(nilStoreBoot), args: []string{"default", "--force"}},
		{name: "import", cmd: newImportCmd(nilStoreBoot), args: []string{"config.json"}},
		{name: "export", cmd: newExportCmd(nilStoreBoot), args: []string{"default"}},
	}
	for _, tc := range nilStoreCases {
		t.Run(tc.name+" storage unavailable", func(t *testing.T) {
			_, _, err := executeConfigProfileCommand(t, tc.cmd, "", tc.args...)
			require.Error(t, err)
			assert.ErrorContains(t, err, "bootstrap: config profile storage unavailable")
		})
	}

	t.Run("list store error", func(t *testing.T) {
		cmd := newListCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{listErr: errors.New("list failed")}, nil)}, nil
		})
		_, _, err := executeConfigProfileCommand(t, cmd, "")
		require.Error(t, err)
		assert.ErrorContains(t, err, "list profiles: list failed")
	})

	t.Run("create validation and store branches", func(t *testing.T) {
		cmd := newCreateCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{}, nil)}, nil
		})
		_, _, err := executeConfigProfileCommand(t, cmd, "", "new", "--preset", "unknown")
		require.Error(t, err)
		assert.ErrorContains(t, err, `unknown preset "unknown"`)

		cmd = newCreateCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{existsErr: errors.New("exists failed")}, nil)}, nil
		})
		_, _, err = executeConfigProfileCommand(t, cmd, "", "new")
		require.Error(t, err)
		assert.ErrorContains(t, err, "check profile: exists failed")

		cmd = newCreateCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{
				Storage: storage.NewFacade(&mutableProfileStore{profiles: map[string]*config.Config{"new": config.DefaultConfig()}}, nil),
			}, nil
		})
		_, _, err = executeConfigProfileCommand(t, cmd, "", "new")
		require.Error(t, err)
		assert.ErrorContains(t, err, `profile "new" already exists`)

		cmd = newCreateCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{saveErr: errors.New("save failed")}, nil)}, nil
		})
		_, _, err = executeConfigProfileCommand(t, cmd, "", "new")
		require.Error(t, err)
		assert.ErrorContains(t, err, "create profile: save failed")

		store := &mutableProfileStore{}
		cmd = newCreateCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(store, nil)}, nil
		})
		stdout, _, err := executeConfigProfileCommand(t, cmd, "", "research", "--preset", "researcher")
		require.NoError(t, err)
		assert.Contains(t, stdout, `Profile "research" created from preset "researcher".`)
		require.Contains(t, store.profiles, "research")
		assert.True(t, store.profiles["research"].Knowledge.Enabled)
		assert.True(t, store.profiles["research"].Graph.Enabled)
		assert.Equal(t, "openai", store.profiles["research"].Embedding.Provider)
		assert.True(t, store.explicit["research"]["knowledge.enabled"])
		assert.True(t, store.explicit["research"]["graph.enabled"])
		assert.True(t, store.explicit["research"]["embedding.provider"])
	})

	t.Run("mutation and IO command errors", func(t *testing.T) {
		cmd := newUseCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{setActiveErr: errors.New("switch failed")}, nil)}, nil
		})
		_, _, err := executeConfigProfileCommand(t, cmd, "", "default")
		require.Error(t, err)
		assert.ErrorContains(t, err, "switch profile: switch failed")

		cmd = newDeleteCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{deleteErr: errors.New("delete failed")}, nil)}, nil
		})
		_, _, err = executeConfigProfileCommand(t, cmd, "", "default", "--force")
		require.Error(t, err)
		assert.ErrorContains(t, err, "delete profile: delete failed")

		cmd = newImportCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{}, nil)}, nil
		})
		_, _, err = executeConfigProfileCommand(t, cmd, "", filepath.Join(t.TempDir(), "missing.json"))
		require.Error(t, err)
		assert.ErrorContains(t, err, "import config:")

		importFile := filepath.Join(t.TempDir(), "save-error.json")
		require.NoError(t, os.WriteFile(importFile, []byte(`{"agent":{"provider":"openai"}}`), 0o600))
		cmd = newImportCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{saveErr: errors.New("save failed")}, nil)}, nil
		})
		_, _, err = executeConfigProfileCommand(t, cmd, "", importFile)
		require.Error(t, err)
		assert.ErrorContains(t, err, `import config: save profile "default": save failed`)

		importFile = filepath.Join(t.TempDir(), "set-active-error.json")
		require.NoError(t, os.WriteFile(importFile, []byte(`{"agent":{"provider":"openai"}}`), 0o600))
		cmd = newImportCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{setActiveErr: errors.New("set active failed")}, nil)}, nil
		})
		_, _, err = executeConfigProfileCommand(t, cmd, "", importFile)
		require.Error(t, err)
		assert.ErrorContains(t, err, `import config: set active profile "default": set active failed`)

		cmd = newExportCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Storage: storage.NewFacade(&mutableProfileStore{loadErr: errors.New("load failed")}, nil)}, nil
		})
		_, _, err = executeConfigProfileCommand(t, cmd, "", "default")
		require.Error(t, err)
		assert.ErrorContains(t, err, "load profile: load failed")
	})

	t.Run("validate reports invalid active config", func(t *testing.T) {
		cmd := newValidateCmd(func() (*bootstrap.Result, error) {
			cfg := config.DefaultConfig()
			cfg.Server.Port = 0
			return &bootstrap.Result{Config: cfg, ProfileName: "broken"}, nil
		})
		_, _, err := executeConfigProfileCommand(t, cmd, "")
		require.Error(t, err)
		assert.ErrorContains(t, err, "validation failed: configuration validation failed")
		assert.ErrorContains(t, err, "invalid port: 0")
	})
}
