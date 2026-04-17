// Package cliboot provides common bootstrap loaders for CLI subcommands.
// Each function satisfies a standard callback signature used by command constructors.
package cliboot

import (
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

// Version is set by cmd/lango to record the app version in bootstrap diagnostics.
var Version string

// BootResult runs bootstrap and returns the full result.
// The caller is responsible for closing the result via boot.Close().
func BootResult() (*bootstrap.Result, error) {
	return bootstrap.Run(bootstrap.Options{
		Version:            Version,
		StartStorageBroker: true,
	})
}

// Config runs bootstrap, returns only the config, and closes the DB client.
func Config() (*config.Config, error) {
	boot, err := bootstrap.Run(bootstrap.Options{
		Version:            Version,
		StartStorageBroker: true,
	})
	if err != nil {
		return nil, err
	}
	defer boot.Close()
	return boot.Config, nil
}
