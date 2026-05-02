package cockpit

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/runledger"
)

const (
	defaultMissionControlMissionLimit  = 6
	defaultMissionControlActivityLimit = 6
)

// MissionControlProjector derives a deterministic Wave 1 Mission Control view.
type MissionControlProjector struct {
	cfg              *config.Config
	sessionKey       string
	metricsCollector *observability.MetricsCollector
	pendingApprovals *PendingApprovalRegistry
	learningBuffer   *LearningSuggestionBuffer
	activityBuffer   *MissionActivityBuffer
	runLedgerStore   RunLedgerReader
	agentRunStore    AgentRunReader
	missionReader    MissionReader
	missionLimit     int
	activityLimit    int
	nowFn            func() time.Time
}

// NewMissionControlProjector creates a deterministic projector from cockpit deps.
func NewMissionControlProjector(deps Deps) *MissionControlProjector {
	return &MissionControlProjector{
		cfg:              deps.Config,
		sessionKey:       deps.SessionKey,
		metricsCollector: deps.MetricsCollector,
		pendingApprovals: deps.PendingApprovals,
		learningBuffer:   deps.LearningBuffer,
		activityBuffer:   deps.ActivityBuffer,
		runLedgerStore:   deps.RunLedgerStore,
		agentRunStore:    deps.AgentRunStore,
		missionReader:    deps.MissionReader,
		missionLimit:     defaultMissionControlMissionLimit,
		activityLimit:    defaultMissionControlActivityLimit,
		nowFn:            time.Now,
	}
}

// Project derives a deterministic Mission Control snapshot from cockpit-owned data.
func (p *MissionControlProjector) Project(taskSnapshots []background.TaskSnapshot) MissionControlSnapshot {
	if p == nil {
		return MissionControlSnapshot{}
	}

	generatedAt := p.now()
	durableMissions, linkedTaskIDs, missionStoreUnavailable, missionDetailsDegraded, runLedgerDegraded, agentRunDegraded := p.projectDurableMissions(taskSnapshots)
	runtimeOverlays, overlayRunLedgerDegraded, overlayAgentRunDegraded := p.projectBackgroundTasks(taskSnapshots, linkedTaskIDs)
	proposedMissions := p.projectLearningSuggestions()
	sort.SliceStable(durableMissions, func(i, j int) bool {
		return compareMissionViews(durableMissions[i], durableMissions[j])
	})
	sort.SliceStable(runtimeOverlays, func(i, j int) bool {
		return compareMissionViews(runtimeOverlays[i], runtimeOverlays[j])
	})
	sort.SliceStable(proposedMissions, func(i, j int) bool {
		return compareMissionViews(proposedMissions[i], proposedMissions[j])
	})
	missions := append(durableMissions, runtimeOverlays...)
	missions = append(missions, proposedMissions...)
	degradedNote := p.buildDegradedNote(
		missionStoreUnavailable,
		missionDetailsDegraded,
		runLedgerDegraded || overlayRunLedgerDegraded,
		agentRunDegraded || overlayAgentRunDegraded,
	)

	activities := p.projectActivities()
	visibleMissions, hiddenMissionCount, missionOverflow := limitMissions(missions, p.missionLimit)
	visibleActivities, hiddenActivityCount, activityOverflow := limitActivities(activities, p.activityLimit)

	decision := p.projectDecision()

	return MissionControlSnapshot{
		Header: HeaderView{
			ActiveAgentSummary:   buildActiveAgentSummary(missions),
			ModelProviderSummary: p.buildModelProviderSummary(),
			PendingDecisionCount: pendingDecisionCount(p.pendingApprovals),
			DegradedNote:         degradedNote,
			ContextSummary:       p.buildContextSummary(),
			MetricsSummary:       p.buildMetricsSummary(),
		},
		Missions:                visibleMissions,
		Decision:                decision,
		Activities:              visibleActivities,
		HiddenMissionCount:      hiddenMissionCount,
		HiddenActivityCount:     hiddenActivityCount,
		MissionOverflowSummary:  missionOverflow,
		ActivityOverflowSummary: activityOverflow,
		Degraded:                degradedNote != "",
		GeneratedAt:             generatedAt,
	}
}

func (p *MissionControlProjector) projectDurableMissions(taskSnapshots []background.TaskSnapshot) ([]MissionView, map[string]struct{}, bool, bool, bool, bool) {
	linkedTaskIDs := make(map[string]struct{})
	if p.missionReader == nil {
		return nil, linkedTaskIDs, true, false, false, false
	}

	ctx := context.Background()
	rows, err := p.missionReader.ListMissionsBySession(ctx, p.sessionKey, max(p.missionLimit*4, 24))
	if err != nil {
		return nil, linkedTaskIDs, true, false, false, false
	}
	if len(rows) == 0 {
		return nil, linkedTaskIDs, false, false, false, false
	}

	tasksByID := make(map[string]background.TaskSnapshot, len(taskSnapshots))
	for _, task := range taskSnapshots {
		tasksByID[strings.TrimSpace(task.ID)] = task
	}

	missions := make([]MissionView, 0, len(rows))
	var runLedgerDegraded bool
	var agentRunDegraded bool
	var missionDetailsDegraded bool
	for _, row := range rows {
		if row == nil {
			continue
		}
		view := missionViewFromDurableRow(row)
		links, err := p.missionReader.ListExecutionLinks(ctx, row.ID.String())
		if err != nil {
			missionDetailsDegraded = true
			missions = append(missions, view)
			continue
		}
		for _, link := range links {
			if link == nil {
				continue
			}
			executionRef := strings.TrimSpace(link.ExecutionRef)
			switch link.ExecutionKind {
			case mission.ExecutionKindTaskOSExecution:
				if executionRef != "" {
					linkedTaskIDs[executionRef] = struct{}{}
				}
				if task, ok := tasksByID[executionRef]; ok {
					enrichDurableMissionFromTask(&view, task)
				}
			}
			if p.runLedgerStore != nil {
				snap, err := p.runLedgerStore.GetRunSnapshot(ctx, executionRef)
				if err != nil {
					runLedgerDegraded = true
				} else if snap != nil {
					enrichDurableMissionFromRunLedger(&view, snap)
				}
			}
			if p.agentRunStore != nil {
				run, err := p.agentRunStore.Get(executionRef)
				if err != nil {
					agentRunDegraded = true
				} else if run != nil {
					enrichDurableMissionFromAgentRun(&view, run)
				}
			}
		}
		missions = append(missions, view)
	}
	return missions, linkedTaskIDs, false, missionDetailsDegraded, runLedgerDegraded, agentRunDegraded
}

func (p *MissionControlProjector) projectBackgroundTasks(taskSnapshots []background.TaskSnapshot, linkedTaskIDs map[string]struct{}) ([]MissionView, bool, bool) {
	if len(taskSnapshots) == 0 {
		return nil, false, false
	}

	ctx := context.Background()
	missions := make([]MissionView, 0, len(taskSnapshots))
	var runLedgerDegraded bool
	var agentRunDegraded bool
	for _, task := range taskSnapshots {
		if originSession := strings.TrimSpace(task.OriginSession); originSession != "" && originSession != strings.TrimSpace(p.sessionKey) {
			continue
		}
		if _, linked := linkedTaskIDs[strings.TrimSpace(task.ID)]; linked {
			continue
		}
		mission := MissionView{
			ID:         "bg:" + strings.TrimSpace(task.ID),
			Kind:       MissionKindActive,
			Status:     missionStatusFromTask(task.StatusText),
			Title:      firstNonEmptyLine(task.Prompt),
			NextAction: nextActionForTask(task.StatusText),
			UpdatedAt:  newestRelevantTaskTime(task),
		}
		if mission.Title == "" {
			mission.Title = strings.TrimSpace(task.ID)
		}

		if p.runLedgerStore != nil {
			snap, err := p.runLedgerStore.GetRunSnapshot(ctx, task.ID)
			if err != nil {
				runLedgerDegraded = true
			} else if snap != nil {
				enrichMissionFromRunLedger(&mission, snap)
			}
		}

		if p.agentRunStore != nil {
			run, err := p.agentRunStore.Get(task.ID)
			if err != nil {
				agentRunDegraded = true
			} else if run != nil {
				enrichMissionFromAgentRun(&mission, run)
			}
		}

		missions = append(missions, mission)
	}
	return missions, runLedgerDegraded, agentRunDegraded
}

func (p *MissionControlProjector) projectDecision() *DecisionView {
	if p.pendingApprovals == nil {
		return nil
	}
	msg := p.pendingApprovals.Latest()
	if msg == nil {
		return nil
	}

	title := strings.TrimSpace(msg.Request.ToolName)
	if summary := firstNonEmptyLine(msg.Request.Summary); summary != "" {
		if title != "" {
			title = title + ": " + summary
		} else {
			title = summary
		}
	}

	return &DecisionView{
		ID:                   strings.TrimSpace(msg.Request.ID),
		Category:             DecisionCategoryApproval,
		Title:                title,
		Reason:               strings.TrimSpace(msg.ViewModel.RuleExplanation),
		EffectText:           strings.TrimSpace(msg.Request.Summary),
		RiskLevel:            strings.TrimSpace(msg.ViewModel.Risk.Level),
		RiskLabel:            strings.TrimSpace(msg.ViewModel.Risk.Label),
		ApproveLabel:         "Approve",
		DenyLabel:            "Deny",
		AllowForSessionLabel: "Allow for session",
		UpdatedAt:            msg.Request.CreatedAt,
	}
}

func (p *MissionControlProjector) projectLearningSuggestions() []MissionView {
	if p.learningBuffer == nil {
		return nil
	}
	items := p.learningBuffer.Snapshot()
	if len(items) == 0 {
		return nil
	}

	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].Timestamp.Equal(items[j].Timestamp) {
			return items[i].Timestamp.After(items[j].Timestamp)
		}
		return items[i].SuggestionID < items[j].SuggestionID
	})

	missions := make([]MissionView, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.ProposedRule)
		if title == "" {
			title = strings.TrimSpace(item.Pattern)
		}
		if title == "" {
			title = "Inspect learning suggestion"
		}
		missions = append(missions, MissionView{
			ID:         "learn:" + strings.TrimSpace(item.SuggestionID),
			Kind:       MissionKindProposed,
			Status:     MissionStatusPrepared,
			Title:      "Apply learning rule: " + title,
			Detail:     strings.TrimSpace(item.Rationale),
			NextAction: "Review suggestion",
			SourceKind: "proposed_learning",
			SourceRef:  strings.TrimSpace(item.SuggestionID),
			UpdatedAt:  item.Timestamp,
		})
	}
	return missions
}

func (p *MissionControlProjector) projectActivities() []ActivityView {
	if p.activityBuffer == nil {
		return nil
	}
	items := p.activityBuffer.Snapshot()
	if len(items) == 0 {
		return nil
	}
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].Timestamp.Equal(items[j].Timestamp) {
			return items[i].Timestamp.After(items[j].Timestamp)
		}
		return items[i].Summary > items[j].Summary
	})

	views := make([]ActivityView, 0, len(items))
	for _, item := range items {
		views = append(views, ActivityView{
			Kind:      item.Kind,
			Summary:   item.Summary,
			Timestamp: item.Timestamp,
		})
	}
	return views
}

func (p *MissionControlProjector) buildDegradedNote(missionStoreUnavailable, missionDetailsDegraded, runLedgerDegraded, agentRunDegraded bool) string {
	var notes []string
	if missionStoreUnavailable {
		notes = append(notes, "Mission store unavailable")
	} else if missionDetailsDegraded {
		notes = append(notes, "Mission details unavailable")
	}
	if p.runLedgerStore == nil || runLedgerDegraded {
		notes = append(notes, "RunLedger unavailable")
	}
	if p.agentRunStore == nil || agentRunDegraded {
		notes = append(notes, "Agent runtime unavailable")
	}
	return strings.Join(notes, "; ")
}

func missionViewFromDurableRow(row *mission.Mission) MissionView {
	view := MissionView{
		ID:         row.ID.String(),
		Kind:       MissionKindActive,
		Status:     missionStatusFromDurable(row.Status),
		Title:      strings.TrimSpace(row.Title),
		SourceKind: strings.TrimSpace(row.SourceKind),
		UpdatedAt:  row.UpdatedAt,
	}
	if row.SourceRef != nil {
		view.SourceRef = strings.TrimSpace(*row.SourceRef)
	}
	if row.Description != nil {
		view.Detail = strings.TrimSpace(*row.Description)
	}
	if row.CurrentBlockedReason != nil {
		view.BlockedReason = strings.TrimSpace(*row.CurrentBlockedReason)
	}
	if row.CurrentDecisionSummary != nil && view.Detail == "" {
		view.Detail = strings.TrimSpace(*row.CurrentDecisionSummary)
	}

	switch row.Status {
	case mission.StatusPrepared:
		view.NextAction = "Start mission"
	case mission.StatusActive:
		view.NextAction = "Continue mission"
	case mission.StatusWaitingDecision:
		view.NextAction = "Resolve pending decision"
		view.RuntimeHint = "Waiting for decision"
	case mission.StatusBlocked:
		view.NextAction = "Resolve blocker"
	case mission.StatusDone:
		view.NextAction = "Review completed output"
	case mission.StatusCancelled:
		view.NextAction = "Restart if still needed"
	}
	return view
}

func missionStatusFromDurable(status mission.Status) MissionStatus {
	switch status {
	case mission.StatusPrepared:
		return MissionStatusPrepared
	case mission.StatusActive:
		return MissionStatusRunning
	case mission.StatusWaitingDecision:
		return MissionStatusPending
	case mission.StatusBlocked:
		return MissionStatusBlocked
	case mission.StatusDone:
		return MissionStatusDone
	case mission.StatusCancelled:
		return MissionStatusCancelled
	default:
		return MissionStatusUnknown
	}
}

func enrichDurableMissionFromTask(missionView *MissionView, task background.TaskSnapshot) {
	if missionView == nil {
		return
	}
	if missionView.Detail == "" {
		missionView.Detail = firstNonEmptyLine(task.Prompt)
	}
	if taskTime := newestRelevantTaskTime(task); taskTime.After(missionView.UpdatedAt) {
		missionView.UpdatedAt = taskTime
	}
}

func enrichDurableMissionFromRunLedger(missionView *MissionView, snap *runledger.RunSnapshot) {
	if missionView == nil || snap == nil {
		return
	}
	if detail := strings.TrimSpace(snap.Goal); detail != "" {
		missionView.Detail = detail
	}
	if blocker := strings.TrimSpace(firstNonEmptyString(snap.CurrentBlocker, snap.TeammateBlockedReason)); blocker != "" {
		missionView.BlockedReason = blocker
	}
	if hint := strings.TrimSpace(snap.TeammateRuntimeCondition); hint != "" && missionView.RuntimeHint == "" {
		missionView.RuntimeHint = hint
	}
	if step := snap.NextExecutableStep(); step != nil && missionView.NextAction != "Resolve pending decision" {
		missionView.NextAction = "Next step: " + strings.TrimSpace(step.Goal)
	}
	if snap.UpdatedAt.After(missionView.UpdatedAt) {
		missionView.UpdatedAt = snap.UpdatedAt
	}
}

func enrichDurableMissionFromAgentRun(missionView *MissionView, run *agentrt.AgentRun) {
	if missionView == nil || run == nil {
		return
	}
	if owner := strings.TrimSpace(run.RequestedAgent); owner != "" {
		missionView.OwnerAgent = owner
	}
	if hint := runtimeHintForAgentRun(run.RuntimeCondition, run.BlockedReason); hint != "" {
		missionView.RuntimeHint = hint
	}
	if missionView.BlockedReason == "" {
		missionView.BlockedReason = strings.TrimSpace(run.BlockedReason)
	}
}

func buildActiveAgentSummary(missions []MissionView) string {
	if len(missions) == 0 {
		return ""
	}

	owners := make([]string, 0, len(missions))
	seen := make(map[string]struct{}, len(missions))
	for _, mission := range missions {
		if mission.Kind != MissionKindActive {
			continue
		}
		if mission.Status == MissionStatusDone || mission.Status == MissionStatusFailed || mission.Status == MissionStatusCancelled {
			continue
		}
		owner := strings.TrimSpace(mission.OwnerAgent)
		if owner == "" {
			continue
		}
		if _, ok := seen[owner]; ok {
			continue
		}
		seen[owner] = struct{}{}
		owners = append(owners, owner)
	}
	if len(owners) == 0 {
		return ""
	}
	if len(owners) == 1 {
		primary := owners[0]
		return primary + " active"
	}
	return fmt.Sprintf("%s +%d more active", owners[0], len(owners)-1)
}

func (p *MissionControlProjector) buildModelProviderSummary() string {
	if p.cfg == nil {
		return ""
	}
	provider := strings.TrimSpace(p.cfg.Agent.Provider)
	model := strings.TrimSpace(p.cfg.Agent.Model)
	switch {
	case provider != "" && model != "":
		return provider + " / " + model
	case model != "":
		return model
	default:
		return provider
	}
}

func (p *MissionControlProjector) buildContextSummary() string {
	return ""
}

func (p *MissionControlProjector) buildMetricsSummary() string {
	if p.metricsCollector == nil || strings.TrimSpace(p.sessionKey) == "" {
		return ""
	}
	metrics := p.metricsCollector.SessionMetrics(p.sessionKey)
	if metrics == nil || metrics.RequestCount == 0 {
		return ""
	}
	return fmt.Sprintf("%d tokens across %d requests", metrics.TotalTokens, metrics.RequestCount)
}

func (p *MissionControlProjector) now() time.Time {
	if p.nowFn == nil {
		return time.Now()
	}
	return p.nowFn()
}

func missionStatusFromTask(statusText string) MissionStatus {
	switch strings.ToLower(strings.TrimSpace(statusText)) {
	case "pending":
		return MissionStatusPending
	case "running":
		return MissionStatusRunning
	case "done":
		return MissionStatusDone
	case "failed":
		return MissionStatusFailed
	case "cancelled":
		return MissionStatusCancelled
	default:
		return MissionStatusUnknown
	}
}

func nextActionForTask(statusText string) string {
	switch missionStatusFromTask(statusText) {
	case MissionStatusPending:
		return "Wait for execution slot"
	case MissionStatusRunning:
		return "Monitor background execution"
	case MissionStatusBlocked:
		return "Resolve blocker"
	case MissionStatusDone:
		return "Review completed output"
	case MissionStatusFailed:
		return "Inspect failure and retry"
	case MissionStatusCancelled:
		return "Restart if still needed"
	default:
		return "Inspect task state"
	}
}

func newestRelevantTaskTime(task background.TaskSnapshot) time.Time {
	updated := task.StartedAt
	if task.CompletedAt.After(updated) {
		updated = task.CompletedAt
	}
	if task.NextRetryAt.After(updated) {
		updated = task.NextRetryAt
	}
	return updated
}

func enrichMissionFromRunLedger(mission *MissionView, snap *runledger.RunSnapshot) {
	if mission == nil || snap == nil {
		return
	}

	if detail := strings.TrimSpace(snap.Goal); detail != "" {
		mission.Detail = detail
	}
	if blocker := strings.TrimSpace(firstNonEmptyString(snap.CurrentBlocker, snap.TeammateBlockedReason)); blocker != "" {
		mission.Status = MissionStatusBlocked
		mission.BlockedReason = blocker
		mission.NextAction = "Resolve blocker"
	}
	if hint := strings.TrimSpace(snap.TeammateRuntimeCondition); hint != "" && mission.RuntimeHint == "" {
		mission.RuntimeHint = hint
	}
	if step := snap.NextExecutableStep(); step != nil && mission.Status != MissionStatusBlocked {
		mission.NextAction = "Next step: " + strings.TrimSpace(step.Goal)
	}
	if snap.UpdatedAt.After(mission.UpdatedAt) {
		mission.UpdatedAt = snap.UpdatedAt
	}
}

func enrichMissionFromAgentRun(mission *MissionView, run *agentrt.AgentRun) {
	if mission == nil || run == nil {
		return
	}
	if owner := strings.TrimSpace(run.RequestedAgent); owner != "" {
		mission.OwnerAgent = owner
	}
	if hint := runtimeHintForAgentRun(run.RuntimeCondition, run.BlockedReason); hint != "" {
		mission.RuntimeHint = hint
	}
	if reason, nextAction, blocked := blockedStateForAgentRun(run.RuntimeCondition, run.BlockedReason); blocked {
		mission.Status = MissionStatusBlocked
		mission.BlockedReason = reason
		mission.NextAction = nextAction
	}
}

func runtimeHintForAgentRun(condition agentrt.AgentRunCondition, blockedReason string) string {
	switch condition {
	case agentrt.AgentRunConditionBlockedWaitingApproval:
		return "Waiting for approval"
	case agentrt.AgentRunConditionBlockedWaitingMessage:
		return "Waiting for message"
	case agentrt.AgentRunConditionWaitingOnTeammate:
		return "Waiting on teammate"
	case agentrt.AgentRunConditionResuming:
		return "Resuming"
	case agentrt.AgentRunConditionOrphaned:
		return "Orphaned"
	case agentrt.AgentRunConditionRecovering:
		return "Recovering"
	}
	return strings.TrimSpace(blockedReason)
}

func blockedStateForAgentRun(condition agentrt.AgentRunCondition, blockedReason string) (string, string, bool) {
	switch condition {
	case agentrt.AgentRunConditionBlockedWaitingApproval:
		reason := firstNonEmptyString(blockedReason, "Waiting for approval")
		return reason, "Resolve approval request", true
	case agentrt.AgentRunConditionBlockedWaitingMessage:
		reason := firstNonEmptyString(blockedReason, "Waiting for message")
		return reason, "Respond to required message", true
	default:
		return "", "", false
	}
}

func compareMissionViews(left, right MissionView) bool {
	if missionKindPriority(left.Kind) != missionKindPriority(right.Kind) {
		return missionKindPriority(left.Kind) < missionKindPriority(right.Kind)
	}
	if missionStatusPriority(left.Status) != missionStatusPriority(right.Status) {
		return missionStatusPriority(left.Status) < missionStatusPriority(right.Status)
	}
	if !left.UpdatedAt.Equal(right.UpdatedAt) {
		return left.UpdatedAt.After(right.UpdatedAt)
	}
	return left.ID < right.ID
}

func missionKindPriority(kind MissionKind) int {
	switch kind {
	case MissionKindActive:
		return 0
	case MissionKindProposed:
		return 1
	default:
		return 2
	}
}

func missionStatusPriority(status MissionStatus) int {
	switch status {
	case MissionStatusRunning:
		return 0
	case MissionStatusBlocked:
		return 1
	case MissionStatusPrepared:
		return 2
	case MissionStatusPending:
		return 3
	case MissionStatusDone:
		return 4
	case MissionStatusFailed:
		return 5
	case MissionStatusCancelled:
		return 6
	default:
		return 7
	}
}

func limitMissions(items []MissionView, limit int) ([]MissionView, int, string) {
	if limit <= 0 || len(items) <= limit {
		return items, 0, ""
	}
	hidden := len(items) - limit
	return items[:limit], hidden, pluralSummary(hidden, "mission", "missions")
}

func limitActivities(items []ActivityView, limit int) ([]ActivityView, int, string) {
	if limit <= 0 || len(items) <= limit {
		return items, 0, ""
	}
	hidden := len(items) - limit
	return items[:limit], hidden, pluralSummary(hidden, "activity item", "activity items")
}

func pluralSummary(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d more %s", count, singular)
	}
	return fmt.Sprintf("%d more %s", count, plural)
}

func pendingDecisionCount(registry *PendingApprovalRegistry) int {
	if registry == nil || !registry.HasPending() {
		return 0
	}
	return 1
}

func firstNonEmptyLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
