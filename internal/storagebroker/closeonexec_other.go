//go:build !unix

package storagebroker

func markFDCloseOnExec(uintptr) {}
