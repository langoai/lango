package settings

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestNewEditorForEmbedding_SkipsWelcome(t *testing.T) {
	cfg := config.DefaultConfig()
	e := NewEditorForEmbedding(cfg, func(_ *config.Config, _ map[string]bool) error {
		return nil
	})

	assert.Equal(t, StepMenu, e.step)
	assert.NotNil(t, e.OnSave)
}

func TestEmbeddedSave_CallsOnSave(t *testing.T) {
	cfg := config.DefaultConfig()

	var called bool
	var receivedCfg *config.Config
	var receivedKeys map[string]bool

	e := NewEditorForEmbedding(cfg, func(c *config.Config, dk map[string]bool) error {
		called = true
		receivedCfg = c
		receivedKeys = dk
		return nil
	})

	// Mark a field dirty to verify it propagates
	e.state.MarkDirty("agent")

	cmd := e.handleMenuSelection("save")

	assert.True(t, called, "OnSave callback should have been called")
	assert.NotNil(t, receivedCfg)
	assert.True(t, receivedKeys["agent"], "dirty key 'agent' should be in dirtyKeys")
	assert.Nil(t, cmd, "embedded save should not return a command")
	assert.False(t, e.Completed, "Completed should remain false in embedded mode")
	assert.True(t, e.saveSuccess, "saveSuccess should be true after successful save")
}

func TestEmbeddedSave_ErrorSetsErr(t *testing.T) {
	cfg := config.DefaultConfig()
	wantErr := fmt.Errorf("disk full")

	e := NewEditorForEmbedding(cfg, func(_ *config.Config, _ map[string]bool) error {
		return wantErr
	})

	cmd := e.handleMenuSelection("save")

	assert.Nil(t, cmd)
	assert.False(t, e.saveSuccess)
	assert.Equal(t, wantErr, e.err)
}

func TestStandaloneSave_Unchanged(t *testing.T) {
	cfg := config.DefaultConfig()
	e := NewEditorWithConfig(cfg)

	cmd := e.handleMenuSelection("save")

	assert.True(t, e.Completed, "standalone save should set Completed")
	require.NotNil(t, cmd, "standalone save should return tea.Quit")
}

func TestEmbeddedSave_BannerClearsOnNextKey(t *testing.T) {
	cfg := config.DefaultConfig()
	e := NewEditorForEmbedding(cfg, func(_ *config.Config, _ map[string]bool) error {
		return nil
	})

	// Trigger a successful save
	e.handleMenuSelection("save")
	assert.True(t, e.saveSuccess)
	assert.Nil(t, e.err)

	// Simulate a key press — banner flags should be cleared
	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	assert.False(t, e.saveSuccess, "saveSuccess should be cleared on next key")
	assert.Nil(t, e.err)
}

func TestEmbeddedSave_ErrorBannerClearsOnNextKey(t *testing.T) {
	cfg := config.DefaultConfig()
	e := NewEditorForEmbedding(cfg, func(_ *config.Config, _ map[string]bool) error {
		return fmt.Errorf("write error")
	})

	// Trigger a failed save
	e.handleMenuSelection("save")
	assert.NotNil(t, e.err)
	assert.False(t, e.saveSuccess)

	// Simulate a key press — error should be cleared
	e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	assert.Nil(t, e.err, "err should be cleared on next key")
}

func TestEmbeddedSave_BannerAppearsInView(t *testing.T) {
	cfg := config.DefaultConfig()
	e := NewEditorForEmbedding(cfg, func(_ *config.Config, _ map[string]bool) error {
		return nil
	})

	// Trigger a successful save
	e.handleMenuSelection("save")

	view := e.View()
	assert.Contains(t, view, "Settings saved")
}

func TestEmbeddedSave_ErrorBannerAppearsInView(t *testing.T) {
	cfg := config.DefaultConfig()
	e := NewEditorForEmbedding(cfg, func(_ *config.Config, _ map[string]bool) error {
		return fmt.Errorf("permission denied")
	})

	// Trigger a failed save
	e.handleMenuSelection("save")

	view := e.View()
	assert.Contains(t, view, "Save failed")
	assert.Contains(t, view, "permission denied")
}
