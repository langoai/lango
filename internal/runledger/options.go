package runledger

// StoreOptions holds optional configuration for RunLedger store implementations.
type StoreOptions struct {
	// AppendHook is called after a journal event is successfully appended.
	// The hook is invoked synchronously — keep it lightweight.
	AppendHook func(JournalEvent)
}

// StoreOption configures a RunLedger store.
type StoreOption func(*StoreOptions)

// WithAppendHook registers a callback invoked after each successful journal append.
// This enables decoupled consumers (e.g., provenance checkpoint creation) to react
// to journal events without modifying the RunLedgerStore interface.
func WithAppendHook(h func(JournalEvent)) StoreOption {
	return func(o *StoreOptions) {
		o.AppendHook = h
	}
}

func applyStoreOptions(opts []StoreOption) StoreOptions {
	var o StoreOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// AppendHookSetter is implemented by concrete store types that support
// post-construction hook registration. Not part of RunLedgerStore.
type AppendHookSetter interface {
	SetAppendHook(func(JournalEvent))
}
