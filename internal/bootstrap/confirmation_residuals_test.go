package bootstrap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type confirmationResidualsErrorWriter struct {
	err error
}

func (w confirmationResidualsErrorWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestConfirmStorePassphraseReturnsPromptWriteError(t *testing.T) {
	origInput := bootstrapConfirmInput
	origOutput := bootstrapConfirmOutput
	t.Cleanup(func() {
		bootstrapConfirmInput = origInput
		bootstrapConfirmOutput = origOutput
	})

	sentinelErr := errors.New("prompt write failed")
	bootstrapConfirmInput = bytes.NewBufferString("yes\n")
	bootstrapConfirmOutput = confirmationResidualsErrorWriter{err: sentinelErr}

	ok, err := confirmStorePassphrase("Store passphrase?")
	assert.False(t, ok)
	require.ErrorIs(t, err, sentinelErr)
}
