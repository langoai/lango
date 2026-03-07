package browser

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeRodCall_RecoversPanic(t *testing.T) {
	t.Parallel()

	err := safeRodCall(func() error {
		panic("CDP connection lost")
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBrowserPanic)
	assert.Contains(t, err.Error(), "CDP connection lost")
}

func TestSafeRodCall_PassesNormalError(t *testing.T) {
	t.Parallel()

	sentinel := fmt.Errorf("normal error")
	err := safeRodCall(func() error {
		return sentinel
	})
	assert.Equal(t, sentinel, err)
}

func TestSafeRodCall_ReturnsNilOnSuccess(t *testing.T) {
	t.Parallel()

	err := safeRodCall(func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestSafeRodCallValue_RecoversPanic(t *testing.T) {
	t.Parallel()

	val, err := safeRodCallValue(func() (string, error) {
		panic("websocket closed")
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBrowserPanic)
	assert.Empty(t, val)
}

func TestSafeRodCallValue_PassesNormalError(t *testing.T) {
	t.Parallel()

	sentinel := fmt.Errorf("element not found")
	val, err := safeRodCallValue(func() (int, error) {
		return 0, sentinel
	})
	assert.Equal(t, sentinel, err)
	assert.Equal(t, 0, val)
}

func TestSafeRodCallValue_ReturnsValueOnSuccess(t *testing.T) {
	t.Parallel()

	val, err := safeRodCallValue(func() (string, error) {
		return "hello", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestErrBrowserPanic_Unwrap(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("%w: runtime crash", ErrBrowserPanic)
	assert.True(t, errors.Is(wrapped, ErrBrowserPanic))
}
