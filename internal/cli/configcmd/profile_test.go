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
}

func (m *mutableProfileStore) Save(_ context.Context, name string, cfg *config.Config, explicitKeys map[string]bool) error {
	if m.profiles == nil {
		m.profiles = make(map[string]*config.Config)
	}
	m.profiles[name] = cfg
	if m.active == "" {
		m.active = name
	}
	return nil
}

func (m *mutableProfileStore) Load(_ context.Context, name string) (*config.Config, map[string]bool, error) {
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
	if _, ok := m.profiles[name]; !ok {
		return errors.New("not found")
	}
	m.active = name
	return nil
}

func (m *mutableProfileStore) List(_ context.Context) ([]configstore.ProfileInfo, error) {
	return m.infos, nil
}

func (m *mutableProfileStore) Delete(_ context.Context, name string) error {
	delete(m.profiles, name)
	return nil
}

func (m *mutableProfileStore) Exists(_ context.Context, name string) (bool, error) {
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
