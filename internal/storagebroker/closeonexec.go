package storagebroker

import "io"

func markFilesCloseOnExec(files ...io.Closer) {
	for _, file := range files {
		if f, ok := file.(interface{ Fd() uintptr }); ok {
			markFDCloseOnExec(f.Fd())
		}
	}
}
