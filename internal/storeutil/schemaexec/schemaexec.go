package schemaexec

import "sync"

var schemaCreateMu sync.Mutex

// RunExclusive serializes ent schema operations that read or mutate generated
// migration metadata.
func RunExclusive(fn func() error) error {
	schemaCreateMu.Lock()
	defer schemaCreateMu.Unlock()
	return fn()
}
