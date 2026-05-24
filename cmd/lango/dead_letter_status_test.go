package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
)

func TestNewDeadLetterStatusLoader_WrapsBootError(t *testing.T) {
	loader := newDeadLetterStatusLoader(func() (*bootstrap.Result, error) {
		return nil, errors.New("no config")
	})

	bridge, cleanup, err := loader()

	require.Error(t, err)
	assert.Nil(t, bridge)
	assert.Nil(t, cleanup)
	assert.Contains(t, err.Error(), "bootstrap: no config")
}
