package eventbus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventNameCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event Event
		want  string
	}{
		{name: "content saved", event: ContentSavedEvent{}, want: "content.saved"},
		{name: "triples extracted", event: TriplesExtractedEvent{}, want: "triples.extracted"},
		{name: "turn completed", event: TurnCompletedEvent{}, want: "turn.completed"},
		{name: "reputation changed", event: ReputationChangedEvent{}, want: "reputation.changed"},
		{name: "memory graph", event: MemoryGraphEvent{}, want: "memory.graph"},
		{name: "tool execution paid", event: ToolExecutionPaidEvent{}, want: "tool.execution.paid"},
		{name: "agent discovered", event: AgentDiscoveredEvent{}, want: "agent.discovered"},
		{name: "task delegated", event: TaskDelegatedEvent{}, want: "task.delegated"},
		{name: "task completed", event: TaskCompletedEvent{}, want: "task.completed"},
		{name: "task failed", event: TaskFailedEvent{}, want: "task.failed"},
		{name: "payment negotiated", event: PaymentNegotiatedEvent{}, want: "payment.negotiated"},
		{name: "payment settled", event: PaymentSettledEvent{}, want: "payment.settled"},
		{name: "trust updated", event: TrustUpdatedEvent{}, want: "trust.updated"},
		{name: "schema exchanged", event: SchemaExchangeEvent{}, want: "schema.exchanged"},
		{name: "policy decision", event: PolicyDecisionEvent{}, want: "policy.decision"},
		{name: "runledger mirror failure", event: RunLedgerMirrorFailureEvent{}, want: "runledger.mirror.failure"},
		{name: "alert triggered", event: AlertEvent{}, want: "alert.triggered"},
		{name: "channel message received", event: ChannelMessageReceivedEvent{}, want: "channel.message.received"},
		{name: "channel message sent", event: ChannelMessageSentEvent{}, want: "channel.message.sent"},
		{name: "sandbox decision", event: SandboxDecisionEvent{}, want: "sandbox.decision"},
		{name: "budget alert", event: BudgetAlertEvent{}, want: "budget.alert"},
		{name: "budget exhausted", event: BudgetExhaustedEvent{}, want: "budget.exhausted"},
		{name: "negotiation started", event: NegotiationStartedEvent{}, want: "negotiation.started"},
		{name: "negotiation completed", event: NegotiationCompletedEvent{}, want: "negotiation.completed"},
		{name: "negotiation failed", event: NegotiationFailedEvent{}, want: "negotiation.failed"},
		{name: "escrow created", event: EscrowCreatedEvent{}, want: "escrow.created"},
		{name: "escrow milestone", event: EscrowMilestoneEvent{}, want: "escrow.milestone"},
		{name: "escrow released", event: EscrowReleasedEvent{}, want: "escrow.released"},
		{name: "escrow onchain deposit", event: EscrowOnChainDepositEvent{}, want: "escrow.onchain.deposit"},
		{name: "escrow onchain work", event: EscrowOnChainWorkEvent{}, want: "escrow.onchain.work"},
		{name: "escrow onchain release", event: EscrowOnChainReleaseEvent{}, want: "escrow.onchain.release"},
		{name: "escrow onchain refund", event: EscrowOnChainRefundEvent{}, want: "escrow.onchain.refund"},
		{name: "escrow onchain dispute", event: EscrowOnChainDisputeEvent{}, want: "escrow.onchain.dispute"},
		{name: "escrow onchain resolved", event: EscrowOnChainResolvedEvent{}, want: "escrow.onchain.resolved"},
		{name: "escrow reorg detected", event: EscrowReorgDetectedEvent{}, want: "escrow.reorg.detected"},
		{name: "escrow dangling", event: EscrowDanglingEvent{}, want: "escrow.dangling"},
		{name: "workspace created", event: WorkspaceCreatedEvent{}, want: "workspace.created"},
		{name: "workspace member joined", event: WorkspaceMemberJoinedEvent{}, want: "workspace.member.joined"},
		{name: "workspace member left", event: WorkspaceMemberLeftEvent{}, want: "workspace.member.left"},
		{name: "workspace commit received", event: WorkspaceCommitReceivedEvent{}, want: "workspace.commit.received"},
		{name: "workspace message posted", event: WorkspaceMessagePostedEvent{}, want: "workspace.message.posted"},
		{name: "workspace archived", event: WorkspaceArchivedEvent{}, want: "workspace.archived"},
		{name: "workspace git divergence", event: WorkspaceGitDivergenceEvent{}, want: "workspace.git.divergence"},
		{name: "context injected", event: ContextInjectedEvent{}, want: "context.injected"},
		{name: "mode changed", event: ModeChangedEvent{}, want: "session.mode.changed"},
		{name: "token usage", event: TokenUsageEvent{}, want: "token.usage"},
		{name: "graph admission batch", event: GraphAdmissionBatchEvent{}, want: "graph.admission.batch"},
		{name: "graph admission unmapped source", event: GraphAdmissionUnmappedSourceEvent{}, want: "graph.unmapped_source"},
		{name: "graph extractor dropped unknown", event: GraphExtractorDroppedUnknownEvent{}, want: "graph.extractor.dropped_unknown"},
		{name: "graph admission write failure", event: GraphAdmissionWriteFailureEvent{}, want: "graph.write_failure"},
		{name: "team formed", event: TeamFormedEvent{}, want: "team.formed"},
		{name: "team disbanded", event: TeamDisbandedEvent{}, want: "team.disbanded"},
		{name: "team member joined", event: TeamMemberJoinedEvent{}, want: "team.member.joined"},
		{name: "team member left", event: TeamMemberLeftEvent{}, want: "team.member.left"},
		{name: "team task delegated", event: TeamTaskDelegatedEvent{}, want: "team.task.delegated"},
		{name: "team task completed", event: TeamTaskCompletedEvent{}, want: "team.task.completed"},
		{name: "team conflict detected", event: TeamConflictDetectedEvent{}, want: "team.conflict.detected"},
		{name: "team payment agreed", event: TeamPaymentAgreedEvent{}, want: "team.payment.agreed"},
		{name: "team health check", event: TeamHealthCheckEvent{}, want: "team.health.check"},
		{name: "team leader changed", event: TeamLeaderChangedEvent{}, want: "team.leader.changed"},
		{name: "team member unhealthy", event: TeamMemberUnhealthyEvent{}, want: "team.member.unhealthy"},
		{name: "team budget warning", event: TeamBudgetWarningEvent{}, want: "team.budget.warning"},
		{name: "team graceful shutdown", event: TeamGracefulShutdownEvent{}, want: "team.graceful.shutdown"},
		{name: "compaction completed", event: CompactionCompletedEvent{}, want: "compaction.completed"},
		{name: "compaction slow", event: CompactionSlowEvent{}, want: "compaction.slow"},
		{name: "learning suggestion", event: LearningSuggestionEvent{}, want: "learning.suggestion"},
		{name: "spec drift detected", event: SpecDriftDetectedEvent{}, want: "learning.spec_drift"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.event.EventName())
		})
	}
}

func TestPublishSandboxDecision(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		PublishSandboxDecision(nil, SandboxDecisionEvent{Source: "exec"})
	})

	bus := New()
	var got SandboxDecisionEvent
	SubscribeTyped(bus, func(evt SandboxDecisionEvent) {
		got = evt
	})

	before := time.Now()
	PublishSandboxDecision(bus, SandboxDecisionEvent{
		SessionKey: "session-1",
		Source:     "exec",
		Command:    "go test ./internal/eventbus",
		Decision:   "applied",
		Backend:    "seatbelt",
	})

	assert.Equal(t, "session-1", got.SessionKey)
	assert.Equal(t, "exec", got.Source)
	assert.Equal(t, "go test ./internal/eventbus", got.Command)
	assert.Equal(t, "applied", got.Decision)
	assert.Equal(t, "seatbelt", got.Backend)
	assert.False(t, got.Timestamp.IsZero())
	assert.False(t, got.Timestamp.Before(before))
	assert.Equal(t, "sandbox.decision", got.EventName())
}
