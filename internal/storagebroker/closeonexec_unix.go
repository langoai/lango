//go:build unix

package storagebroker

import "golang.org/x/sys/unix"

func markFDCloseOnExec(fd uintptr) {
	unix.CloseOnExec(int(fd))
}
