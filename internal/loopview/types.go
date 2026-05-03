package loopview

import "time"

type LoopKind string

const (
	LoopKindMissionCluster      LoopKind = "mission_cluster"
	LoopKindInquiry             LoopKind = "inquiry"
	LoopKindDeadLetter          LoopKind = "dead_letter"
	LoopKindFollowUp            LoopKind = "follow_up"
	LoopKindScheduledAutomation LoopKind = "scheduled_automation"
)

type LoopStatus string

const (
	LoopStatusActive      LoopStatus = "active"
	LoopStatusWaitingUser LoopStatus = "waiting_user"
	LoopStatusScheduled   LoopStatus = "scheduled"
	LoopStatusBlocked     LoopStatus = "blocked"
	LoopStatusNeedsReview LoopStatus = "needs_review"
	LoopStatusResolved    LoopStatus = "resolved"
)

const (
	DefaultMissionReviewThreshold = 24 * time.Hour
	DefaultInquiryFollowUpAge     = 24 * time.Hour
)

type LoopView struct {
	LoopID     string
	SessionKey string
	LoopKind   LoopKind
	Title      string
	Summary    string
	Status     LoopStatus
	Priority   int
	SourceRefs []string
	NextAction string
	UpdatedAt  time.Time
}

type AgendaView struct {
	SessionKey  string
	GeneratedAt time.Time
	Loops       []LoopView
}

type MissionSource struct {
	MissionID          string
	SessionKey         string
	Title              string
	Status             string
	UpdatedAt          time.Time
	NeedsReview        bool
	HasActiveExecution bool
}

type ProposalSource struct {
	ProposalID         string
	SessionKey         string
	Title              string
	Status             string
	UpdatedAt          time.Time
	HasActiveExecution bool
}

type InquirySource struct {
	InquiryID  string
	SessionKey string
	Topic      string
	Question   string
	CreatedAt  time.Time
}

type DeadLetterSource struct {
	ReferenceID string
	Title       string
	Summary     string
	Retryable   bool
	UpdatedAt   time.Time
}

type CronSource struct {
	JobID         string
	Name          string
	Enabled       bool
	NextRunAt     time.Time
	LastRunStatus string
	LastRunAt     time.Time
}

type ProjectionInput struct {
	SessionKey  string
	Missions    []MissionSource
	Proposals   []ProposalSource
	Inquiries   []InquirySource
	DeadLetters []DeadLetterSource
	CronJobs    []CronSource
}
