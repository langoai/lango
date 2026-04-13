package agent

// ExposurePolicy controls how a tool appears in agent prompts.
// Zero value (ExposureDefault) means the tool is always visible (backward compatible).
type ExposurePolicy int

const (
	// ExposureDefault means the tool is always visible (backward compatible zero value).
	ExposureDefault ExposurePolicy = iota
	// ExposureAlwaysVisible forces the tool into the prompt.
	ExposureAlwaysVisible
	// ExposureDeferred means the tool is only discoverable via builtin_search.
	ExposureDeferred
	// ExposureHidden means the tool is never shown in prompts or search.
	ExposureHidden
)

// String returns the human-readable name of the exposure policy.
func (e ExposurePolicy) String() string {
	switch e {
	case ExposureDefault:
		return "default"
	case ExposureAlwaysVisible:
		return "always_visible"
	case ExposureDeferred:
		return "deferred"
	case ExposureHidden:
		return "hidden"
	default:
		return "default"
	}
}

// IsVisible reports whether the tool should appear in default prompt listings.
func (e ExposurePolicy) IsVisible() bool {
	return e == ExposureDefault || e == ExposureAlwaysVisible
}

// ActivityKind classifies what a tool does at a high level.
type ActivityKind string

const (
	ActivityRead    ActivityKind = "read"
	ActivityWrite   ActivityKind = "write"
	ActivityExecute ActivityKind = "execute"
	ActivityQuery   ActivityKind = "query"
	ActivityManage  ActivityKind = "manage"
)

// ToolCapability holds rich metadata for tool discovery and policy enforcement.
// All fields have zero-value defaults that preserve existing behavior.
type ToolCapability struct {
	// Aliases are alternate names that match during tool search (e.g. "ls" for "fs_list").
	Aliases []string
	// Category is a semantic category hint (e.g. "filesystem", "crypto").
	// Distinct from the Catalog category — this is tool-level metadata.
	Category string
	// SearchHints are additional keywords for search ranking.
	SearchHints []string
	// Exposure controls prompt visibility policy.
	// Zero value (ExposureDefault) = always visible = backward compatible.
	Exposure ExposurePolicy
	// ReadOnly indicates the tool performs no mutations.
	// Zero value (false) = assume mutation possible = fail-safe.
	ReadOnly bool
	// ConcurrencySafe indicates the tool can be called concurrently.
	// Zero value (false) = assume not safe for concurrency = fail-safe.
	ConcurrencySafe bool
	// Activity classifies the tool's primary action.
	Activity ActivityKind
	// RequiredCapabilities lists system capabilities needed (e.g. "payment", "encryption").
	RequiredCapabilities []string
}
