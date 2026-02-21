package types

// RPCSenderFunc defines how to send an RPC request to a remote provider.
type RPCSenderFunc func(event string, payload interface{}) error
