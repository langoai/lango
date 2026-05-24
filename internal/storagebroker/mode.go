package storagebroker

import "os"

const brokerFlag = "--storage-broker"

// IsBrokerMode reports whether the current process should run as the storage broker.
func IsBrokerMode() bool {
	for _, arg := range os.Args[1:] {
		if arg == brokerFlag {
			return true
		}
	}
	return false
}
