package loopview

import (
	"slices"
	"strings"
	"time"
)

type Projector struct {
	nowFn                  func() time.Time
	missionReviewThreshold time.Duration
	inquiryFollowUpAge     time.Duration
}

func NewProjector(nowFn func() time.Time) *Projector {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Projector{
		nowFn:                  nowFn,
		missionReviewThreshold: DefaultMissionReviewThreshold,
		inquiryFollowUpAge:     DefaultInquiryFollowUpAge,
	}
}

func (p *Projector) Project(in ProjectionInput) AgendaView {
	now := p.nowFn()

	loops := make([]LoopView, 0,
		len(in.Missions)+len(in.Inquiries)+len(in.DeadLetters)+len(in.CronJobs)+len(in.Proposals),
	)
	loops = append(loops, p.projectMissionLoops(in.SessionKey, in.Missions)...)
	loops = append(loops, p.projectInquiryLoops(in.SessionKey, in.Inquiries)...)
	loops = append(loops, p.projectDeadLetterLoops(in.DeadLetters)...)
	loops = append(loops, p.projectCronLoops(in.CronJobs)...)
	loops = append(loops, p.projectFollowUpLoops(in.SessionKey, now, in.Missions, in.Proposals, in.Inquiries)...)

	sortLoops(loops)
	return AgendaView{
		SessionKey:  in.SessionKey,
		GeneratedAt: now,
		Loops:       loops,
	}
}

func (p *Projector) projectMissionLoops(sessionKey string, missions []MissionSource) []LoopView {
	out := make([]LoopView, 0, len(missions))
	for _, mission := range missions {
		if strings.TrimSpace(mission.SessionKey) != strings.TrimSpace(sessionKey) {
			continue
		}
		out = append(out, LoopView{
			LoopID:     "mission:" + mission.MissionID,
			SessionKey: mission.SessionKey,
			LoopKind:   LoopKindMissionCluster,
			Title:      mission.Title,
			Summary:    missionStatusSummary(mission.Status),
			Status:     missionToLoopStatus(mission.Status),
			Priority:   statusPriority(missionToLoopStatus(mission.Status)),
			SourceRefs: []string{mission.MissionID},
			NextAction: missionNextAction(mission.Status),
			UpdatedAt:  mission.UpdatedAt,
		})
	}
	return out
}

func (p *Projector) projectInquiryLoops(sessionKey string, inquiries []InquirySource) []LoopView {
	out := make([]LoopView, 0, len(inquiries))
	for _, inquiry := range inquiries {
		if strings.TrimSpace(inquiry.SessionKey) != strings.TrimSpace(sessionKey) {
			continue
		}
		title := strings.TrimSpace(inquiry.Topic)
		if title == "" {
			title = "Pending inquiry"
		}
		out = append(out, LoopView{
			LoopID:     "inquiry:" + inquiry.InquiryID,
			SessionKey: inquiry.SessionKey,
			LoopKind:   LoopKindInquiry,
			Title:      title,
			Summary:    inquiry.Question,
			Status:     LoopStatusWaitingUser,
			Priority:   statusPriority(LoopStatusWaitingUser),
			SourceRefs: []string{inquiry.InquiryID},
			NextAction: "Answer inquiry",
			UpdatedAt:  inquiry.CreatedAt,
		})
	}
	return out
}

func (p *Projector) projectDeadLetterLoops(items []DeadLetterSource) []LoopView {
	out := make([]LoopView, 0, len(items))
	for _, item := range items {
		status := LoopStatusResolved
		nextAction := "Review dead-letter history"
		if item.Retryable {
			status = LoopStatusBlocked
			nextAction = "Retry or manually replay"
		}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = "Dead-letter backlog item"
		}
		out = append(out, LoopView{
			LoopID:     "dead-letter:" + item.ReferenceID,
			SessionKey: "",
			LoopKind:   LoopKindDeadLetter,
			Title:      title,
			Summary:    item.Summary,
			Status:     status,
			Priority:   statusPriority(status),
			SourceRefs: []string{item.ReferenceID},
			NextAction: nextAction,
			UpdatedAt:  item.UpdatedAt,
		})
	}
	return out
}

func (p *Projector) projectCronLoops(items []CronSource) []LoopView {
	out := make([]LoopView, 0, len(items))
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		status := LoopStatusScheduled
		nextAction := "Wait for scheduled run"
		updatedAt := item.NextRunAt
		summary := "Scheduled automation"
		if strings.EqualFold(strings.TrimSpace(item.LastRunStatus), "failed") {
			status = LoopStatusBlocked
			nextAction = "Review failed cron run"
			updatedAt = item.LastRunAt
			summary = "Scheduled automation failed"
		}
		out = append(out, LoopView{
			LoopID:     "cron:" + item.JobID,
			SessionKey: "",
			LoopKind:   LoopKindScheduledAutomation,
			Title:      item.Name,
			Summary:    summary,
			Status:     status,
			Priority:   statusPriority(status),
			SourceRefs: []string{item.JobID},
			NextAction: nextAction,
			UpdatedAt:  updatedAt,
		})
	}
	return out
}

func (p *Projector) projectFollowUpLoops(sessionKey string, now time.Time, missions []MissionSource, proposals []ProposalSource, inquiries []InquirySource) []LoopView {
	out := make([]LoopView, 0)

	for _, proposal := range proposals {
		if strings.TrimSpace(proposal.SessionKey) != strings.TrimSpace(sessionKey) {
			continue
		}
		if strings.TrimSpace(proposal.Status) != "accepted" || proposal.HasActiveExecution {
			continue
		}
		out = append(out, LoopView{
			LoopID:     "follow-up:proposal:" + proposal.ProposalID,
			SessionKey: proposal.SessionKey,
			LoopKind:   LoopKindFollowUp,
			Title:      proposal.Title,
			Summary:    "Accepted proposal has no active execution yet",
			Status:     LoopStatusActive,
			Priority:   statusPriority(LoopStatusActive),
			SourceRefs: []string{proposal.ProposalID},
			NextAction: "Start linked execution",
			UpdatedAt:  proposal.UpdatedAt,
		})
	}

	for _, mission := range missions {
		if strings.TrimSpace(mission.SessionKey) != strings.TrimSpace(sessionKey) {
			continue
		}
		if strings.TrimSpace(mission.Status) != "done" || !mission.NeedsReview {
			continue
		}
		if mission.UpdatedAt.Add(p.missionReviewThreshold).Before(now) {
			continue
		}
		out = append(out, LoopView{
			LoopID:     "follow-up:mission:" + mission.MissionID,
			SessionKey: mission.SessionKey,
			LoopKind:   LoopKindFollowUp,
			Title:      mission.Title,
			Summary:    "Completed mission still needs review",
			Status:     LoopStatusNeedsReview,
			Priority:   statusPriority(LoopStatusNeedsReview),
			SourceRefs: []string{mission.MissionID},
			NextAction: "Review completed mission",
			UpdatedAt:  mission.UpdatedAt,
		})
	}

	for _, inquiry := range inquiries {
		if strings.TrimSpace(inquiry.SessionKey) != strings.TrimSpace(sessionKey) {
			continue
		}
		if inquiry.CreatedAt.Add(p.inquiryFollowUpAge).After(now) {
			continue
		}
		title := strings.TrimSpace(inquiry.Topic)
		if title == "" {
			title = "Pending inquiry follow-up"
		}
		out = append(out, LoopView{
			LoopID:     "follow-up:inquiry:" + inquiry.InquiryID,
			SessionKey: inquiry.SessionKey,
			LoopKind:   LoopKindFollowUp,
			Title:      title,
			Summary:    "Pending inquiry is aging without user input",
			Status:     LoopStatusNeedsReview,
			Priority:   statusPriority(LoopStatusNeedsReview),
			SourceRefs: []string{inquiry.InquiryID},
			NextAction: "Follow up on inquiry",
			UpdatedAt:  inquiry.CreatedAt,
		})
	}

	return out
}

func sortLoops(loops []LoopView) {
	slices.SortFunc(loops, func(a, b LoopView) int {
		if pa, pb := statusPriority(a.Status), statusPriority(b.Status); pa != pb {
			if pa < pb {
				return -1
			}
			return 1
		}
		switch {
		case a.UpdatedAt.After(b.UpdatedAt):
			return -1
		case a.UpdatedAt.Before(b.UpdatedAt):
			return 1
		case a.Title < b.Title:
			return -1
		case a.Title > b.Title:
			return 1
		case a.LoopID < b.LoopID:
			return -1
		case a.LoopID > b.LoopID:
			return 1
		default:
			return 0
		}
	})
}

func statusPriority(status LoopStatus) int {
	switch status {
	case LoopStatusWaitingUser:
		return 1
	case LoopStatusBlocked:
		return 2
	case LoopStatusActive:
		return 3
	case LoopStatusScheduled:
		return 4
	case LoopStatusNeedsReview:
		return 5
	case LoopStatusResolved:
		return 6
	default:
		return 99
	}
}

func missionToLoopStatus(status string) LoopStatus {
	switch strings.TrimSpace(status) {
	case "waiting_decision":
		return LoopStatusWaitingUser
	case "blocked":
		return LoopStatusBlocked
	case "active", "prepared":
		return LoopStatusActive
	case "done", "cancelled":
		return LoopStatusResolved
	default:
		return LoopStatusNeedsReview
	}
}

func missionStatusSummary(status string) string {
	switch strings.TrimSpace(status) {
	case "waiting_decision":
		return "Waiting on user direction"
	case "blocked":
		return "Blocked work"
	case "active":
		return "Active mission"
	case "prepared":
		return "Prepared mission"
	case "done":
		return "Completed mission"
	case "cancelled":
		return "Cancelled mission"
	default:
		return "Mission"
	}
}

func missionNextAction(status string) string {
	switch strings.TrimSpace(status) {
	case "waiting_decision":
		return "Provide direction"
	case "blocked":
		return "Resolve blocker"
	case "active":
		return "Continue mission"
	case "prepared":
		return "Start mission"
	case "done":
		return "Review completion"
	case "cancelled":
		return "Reopen if needed"
	default:
		return "Review mission"
	}
}
