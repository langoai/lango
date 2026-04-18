//go:build unix

package storagebroker

import (
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestMarkCloseOnExecSetsFlag(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	markFilesCloseOnExec(r, w)

	assertCloseOnExec := func(file *os.File) {
		t.Helper()
		flags, _, errno := unix.Syscall(unix.SYS_FCNTL, file.Fd(), unix.F_GETFD, 0)
		if errno != 0 {
			t.Fatalf("fcntl(F_GETFD) fd=%d: %v", file.Fd(), errno)
		}
		if flags&unix.FD_CLOEXEC == 0 {
			t.Fatalf("expected FD_CLOEXEC for fd=%d", file.Fd())
		}
	}

	assertCloseOnExec(r)
	assertCloseOnExec(w)
}
