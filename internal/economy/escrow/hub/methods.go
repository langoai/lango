package hub

// Contract method name constants for LangoEscrowHub and LangoEscrowHubV2.
const (
	// V1 methods
	MethodCreateDeal     = "createDeal"
	MethodDeposit        = "deposit"
	MethodSubmitWork     = "submitWork"
	MethodRelease        = "release"
	MethodRefund         = "refund"
	MethodDispute        = "dispute"
	MethodResolveDispute = "resolveDispute"
	MethodGetDeal        = "getDeal"
	MethodNextDealID     = "nextDealId"

	// V2-specific methods
	MethodDirectSettle          = "directSettle"
	MethodCreateSimpleEscrow    = "createSimpleEscrow"
	MethodCreateMilestoneEscrow = "createMilestoneEscrow"
	MethodCreateTeamEscrow      = "createTeamEscrow"
	MethodCompleteMilestone     = "completeMilestone"
	MethodReleaseMilestone      = "releaseMilestone"
)
