//go:build unix

package storagebroker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestMarkFilesCloseOnExec(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	clearFDFlags(t, r)
	clearFDFlags(t, w)

	markFilesCloseOnExec(r, w)

	assertCloseOnExec(t, r)
	assertCloseOnExec(t, w)
}

func clearFDFlags(t *testing.T, f *os.File) {
	t.Helper()
	flags, err := unix.FcntlInt(f.Fd(), unix.F_GETFD, 0)
	require.NoError(t, err)
	_, err = unix.FcntlInt(f.Fd(), unix.F_SETFD, flags&^unix.FD_CLOEXEC)
	require.NoError(t, err)
}

func assertCloseOnExec(t *testing.T, f *os.File) {
	t.Helper()
	flags, err := unix.FcntlInt(f.Fd(), unix.F_GETFD, 0)
	require.NoError(t, err)
	assert.NotZero(t, flags&unix.FD_CLOEXEC)
}
