package main

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	clistatus "github.com/langoai/lango/internal/cli/status"
)

func newDeadLetterStatusLoader(bootLoader func() (*bootstrap.Result, error)) clistatus.DeadLetterBridgeLoader {
	return func() (clistatus.DeadLetterBridge, func(), error) {
		boot, err := bootLoader()
		if err != nil {
			return nil, nil, fmt.Errorf("bootstrap: %w", err)
		}
		application, err := app.New(boot, app.WithLocalChat())
		if err != nil {
			_ = boot.Close()
			return nil, nil, fmt.Errorf("build app: %w", err)
		}
		cleanup := func() {
			_ = application.Stop(context.Background())
			_ = boot.Close()
		}
		bridge := clistatus.NewToolCatalogDeadLetterBridge(application.ToolCatalog)
		if !bridge.Ready() {
			cleanup()
			return nil, nil, clistatus.ErrDeadLetterStatusToolsUnavailable
		}
		return bridge, cleanup, nil
	}
}
