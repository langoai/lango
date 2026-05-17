package cliexit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCarriesCodeAndUnwrapsCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("scripted confirmation required")

	err := New(3, cause)

	code, ok := Code(err)
	require.True(t, ok)
	assert.Equal(t, 3, code)
	assert.False(t, Silent(err))
	assert.ErrorIs(t, err, cause)
	assert.Equal(t, "scripted confirmation required", err.Error())
}

func TestSilentCarriesCodeWithoutUserFacingError(t *testing.T) {
	t.Parallel()

	err := NewSilent(3)

	code, ok := Code(err)
	require.True(t, ok)
	assert.Equal(t, 3, code)
	assert.True(t, Silent(err))
	assert.Equal(t, "exit code 3", err.Error())
}

func TestCodeRejectsPlainErrors(t *testing.T) {
	t.Parallel()

	code, ok := Code(errors.New("plain"))

	assert.False(t, ok)
	assert.Zero(t, code)
	assert.False(t, Silent(errors.New("plain")))
}
