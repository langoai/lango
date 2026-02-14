package passphrase

import (
	"os"
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
			// Create a pipe and replace stdin
			r, w, err := os.Pipe()
			require.NoError(t, err)

			origStdin := os.Stdin
			os.Stdin = r
			t.Cleanup(func() { os.Stdin = origStdin })

			_, err = w.WriteString(tt.giveData)
			require.NoError(t, err)
			require.NoError(t, w.Close())

			got, err := ReadStdinPipe()
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
	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	// Close immediately — no data
	require.NoError(t, w.Close())

	_, err = ReadStdinPipe()
	assert.Error(t, err)
}
