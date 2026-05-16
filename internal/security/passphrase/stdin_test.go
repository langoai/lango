package passphrase

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadStdinPipe(t *testing.T) {
	tests := []struct {
		give     string
		giveData string
		wantPass string
		wantErr  bool
	}{
		{
			give:     "valid passphrase with newline",
			giveData: "my-secret-passphrase\n",
			wantPass: "my-secret-passphrase",
		},
		{
			give:     "passphrase with CRLF",
			giveData: "my-secret-passphrase\r\n",
			wantPass: "my-secret-passphrase",
		},
		{
			give:     "empty input",
			giveData: "\n",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ReadStdinPipeFromReader(bytes.NewBufferString(tt.giveData))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPass, got)
		})
	}
}

func TestReadStdinPipe_EmptyPipe(t *testing.T) {
	_, err := ReadStdinPipeFromReader(bytes.NewBuffer(nil))
	assert.Error(t, err)
	assert.EqualError(t, err, "empty passphrase from stdin")
}

func TestReadStdinPipe_NoTrailingNewlineStillWorks(t *testing.T) {
	got, err := ReadStdinPipeFromReader(bytes.NewBufferString("my-secret-passphrase"))
	require.NoError(t, err)
	assert.Equal(t, "my-secret-passphrase", got)
}
