package passphrase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadKeyfile(t *testing.T) {
	tests := []struct {
		give        string
		givePerm    os.FileMode
		giveContent string
		wantPass    string
		wantErr     bool
	}{
		{
			give:        "valid keyfile",
			givePerm:    0600,
			giveContent: "my-secret-passphrase\n",
			wantPass:    "my-secret-passphrase",
		},
		{
			give:        "no trailing newline",
			givePerm:    0600,
			giveContent: "my-secret-passphrase",
			wantPass:    "my-secret-passphrase",
		},
		{
			give:        "trailing whitespace",
			givePerm:    0600,
			giveContent: "my-secret-passphrase\n\r \t",
			wantPass:    "my-secret-passphrase",
		},
		{
			give:        "wrong permissions",
			givePerm:    0644,
			giveContent: "my-secret-passphrase\n",
			wantErr:     true,
		},
		{
			give:        "empty file",
			givePerm:    0600,
			giveContent: "",
			wantErr:     true,
		},
		{
			give:        "whitespace only",
			givePerm:    0600,
			giveContent: "\n\r \t",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "keyfile")

			require.NoError(t, os.WriteFile(path, []byte(tt.giveContent), tt.givePerm))

			got, err := ReadKeyfile(path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPass, got)
		})
	}
}

func TestReadKeyfile_NotFound(t *testing.T) {
	_, err := ReadKeyfile("/nonexistent/path/keyfile")
	assert.Error(t, err)
}

func TestWriteKeyfile(t *testing.T) {
	tests := []struct {
		give     string
		givePass string
		wantErr  bool
	}{
		{
			give:     "valid passphrase",
			givePass: "my-secret-passphrase",
		},
		{
			give:     "passphrase with special chars",
			givePass: "p@ss!word#123$%^",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "subdir", "keyfile")

			err := WriteKeyfile(path, tt.givePass)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify file exists and has correct permissions
			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, requiredPerm, info.Mode().Perm())

			// Verify content
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, tt.givePass+"\n", string(data))

			// Verify parent directory permissions
			dirInfo, err := os.Stat(filepath.Dir(path))
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())
		})
	}
}

func TestWriteKeyfile_ThenRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keyfile")
	passphrase := "roundtrip-test-passphrase"

	require.NoError(t, WriteKeyfile(path, passphrase))

	got, err := ReadKeyfile(path)
	require.NoError(t, err)
	assert.Equal(t, passphrase, got)
}

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		give     string
		givePerm os.FileMode
		wantErr  bool
	}{
		{
			give:     "correct 0600",
			givePerm: 0600,
		},
		{
			give:     "too open 0644",
			givePerm: 0644,
			wantErr:  true,
		},
		{
			give:     "too open 0777",
			givePerm: 0777,
			wantErr:  true,
		},
		{
			give:     "group readable 0640",
			givePerm: 0640,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "keyfile")

			require.NoError(t, os.WriteFile(path, []byte("test"), tt.givePerm))

			err := ValidatePermissions(path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePermissions_NotFound(t *testing.T) {
	err := ValidatePermissions("/nonexistent/path/keyfile")
	assert.Error(t, err)
}

func TestShredKeyfile(t *testing.T) {
	t.Run("shreds existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "keyfile")
		content := "super-secret-passphrase"

		require.NoError(t, os.WriteFile(path, []byte(content), 0600))

		err := ShredKeyfile(path)
		require.NoError(t, err)

		_, statErr := os.Stat(path)
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("returns nil for nonexistent file", func(t *testing.T) {
		err := ShredKeyfile("/nonexistent/path/keyfile")
		assert.NoError(t, err)
	})
}
