package tuicore

// FieldOptionsLoadedMsg is sent when an asynchronous model fetch completes.
// ProviderID is recorded at request time so stale results (from a previously
// selected provider) can be safely ignored.
type FieldOptionsLoadedMsg struct {
	FieldKey   string   // target field Key
	ProviderID string   // provider at request time (race-condition guard)
	Options    []string // fetched options (nil on error)
	Err        error    // fetch error, if any
}
