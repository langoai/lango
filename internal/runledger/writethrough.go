package runledger

// RolloutStage controls how deeply the RunLedger is integrated.
type RolloutStage int

const (
	// StageShadow: journal records only, existing systems unaffected.
	StageShadow RolloutStage = iota
	// StageWriteThrough: all creates/updates go through ledger first, then mirror to projections.
	StageWriteThrough
	// StageAuthoritativeRead: state reads come from ledger snapshots only.
	StageAuthoritativeRead
	// StageProjectionRetired: legacy direct writes removed.
	StageProjectionRetired
)

// RolloutConfig holds the current rollout stage configuration.
type RolloutConfig struct {
	Stage RolloutStage
}

// IsShadow returns true if only shadow journaling is active.
func (c RolloutConfig) IsShadow() bool {
	return c.Stage == StageShadow
}

// IsWriteThrough returns true if write-through is active.
func (c RolloutConfig) IsWriteThrough() bool {
	return c.Stage >= StageWriteThrough
}

// IsAuthoritativeRead returns true if reads should come from ledger.
func (c RolloutConfig) IsAuthoritativeRead() bool {
	return c.Stage >= StageAuthoritativeRead
}
