package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/lifecycle"
)

func TestAppStartPropagatesRegistryErrorAndRollsBackStartedComponents(t *testing.T) {
	t.Parallel()

	startErr := errors.New("startup rejected")
	var events []string
	application := &App{registry: lifecycle.NewRegistry()}
	application.registry.Register(lifecycle.NewFuncComponent(
		"alpha",
		func(context.Context, *sync.WaitGroup) error {
			events = append(events, "start-alpha")
			return nil
		},
		func(context.Context) error {
			events = append(events, "stop-alpha")
			return nil
		},
	), lifecycle.PriorityInfra)
	application.registry.Register(lifecycle.NewFuncComponent(
		"beta",
		func(context.Context, *sync.WaitGroup) error {
			events = append(events, "start-beta")
			return startErr
		},
		func(context.Context) error {
			events = append(events, "stop-beta")
			return nil
		},
	), lifecycle.PriorityCore)

	err := application.Start(context.Background())

	require.Error(t, err)
	assert.ErrorIs(t, err, startErr)
	assert.ErrorContains(t, err, "start beta")
	assert.Equal(t, []string{"start-alpha", "start-beta", "stop-alpha"}, events)
}

func TestAppStopJoinsComponentStopAndWaitErrors(t *testing.T) {
	t.Parallel()

	stopErr := errors.New("stop refused")
	application := &App{registry: lifecycle.NewRegistry()}
	application.registry.Register(lifecycle.NewFuncComponent(
		"stubborn",
		func(context.Context, *sync.WaitGroup) error { return nil },
		func(context.Context) error { return stopErr },
	), lifecycle.PriorityCore)
	require.NoError(t, application.Start(context.Background()))
	application.wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	err := application.Stop(ctx)
	application.wg.Done()

	require.Error(t, err)
	assert.ErrorIs(t, err, stopErr)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestBuildCatalogFromEntriesPreservesEmptyAndDisabledCategories(t *testing.T) {
	t.Parallel()

	catalog := buildCatalogFromEntries([]appinit.CatalogEntry{
		{
			Category:    "disabled",
			Description: "Disabled category",
			ConfigKey:   "feature.disabled",
			Enabled:     false,
		},
		{
			Category:    "empty",
			Description: "Enabled but empty",
			Enabled:     true,
			Tools:       nil,
		},
		{
			Category:    "visible",
			Description: "Visible tools",
			Enabled:     true,
			Tools: []*agent.Tool{{
				Name:        "appStartPropagatesRegistryErrorAndRollsBackStartedComponents_visible",
				Description: "Visible registry tool",
				Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
					return "ok", nil
				},
			}},
		},
	})

	require.NotNil(t, catalog)
	assert.Equal(t, 1, catalog.ToolCount())
	assert.Empty(t, catalog.ToolNamesForCategory("disabled"))
	assert.Empty(t, catalog.ToolNamesForCategory("empty"))
	assert.Equal(t, []string{"appStartPropagatesRegistryErrorAndRollsBackStartedComponents_visible"}, catalog.ToolNamesForCategory("visible"))

	section := (&catalogSourceAdapter{catalog: catalog, cfg: config.DefaultConfig()}).
		BuildToolCatalogSection("")
	assert.Contains(t, section, "**visible** (Visible tools): appStartPropagatesRegistryErrorAndRollsBackStartedComponents_visible")
	assert.Contains(t, section, "Disabled categories (enable via config): disabled (feature.disabled)")
}
