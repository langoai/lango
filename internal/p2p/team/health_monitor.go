package team

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

// HealthMonitor periodically pings team members and publishes unhealthy events
// when a member misses too many consecutive health checks.
type HealthMonitor struct {
	coordinator *Coordinator
	bus         *eventbus.Bus
	logger      *zap.SugaredLogger
	interval    time.Duration
	maxMissed   int
	invokeFn    InvokeFunc

	mu        sync.RWMutex
	missCount map[string]map[string]int // teamID -> memberDID -> consecutive misses
	lastSeen  map[string]map[string]time.Time
	gitState  map[string]map[string]string // workspaceID -> memberDID -> headHash

	gitStateProv GitStateProvider
	workspaceIDs func() []string

	subsOnce sync.Once // ensures event subscriptions are registered only once
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// HealthMonitorConfig configures the health monitor.
type HealthMonitorConfig struct {
	Coordinator      *Coordinator
	Bus              *eventbus.Bus
	Logger           *zap.SugaredLogger
	Interval         time.Duration
	MaxMissed        int
	InvokeFn         InvokeFunc
	GitStateProvider GitStateProvider
	WorkspaceIDsFn   func() []string
}

// GitStateProvider returns the HEAD commit hash for a workspace from a given member.
type GitStateProvider func(ctx context.Context, peerID, workspaceID string) (string, error)

// MemberGitState records a member's known HEAD hash for a workspace.
type MemberGitState struct {
	MemberDID string
	HeadHash  string
	UpdatedAt time.Time
}

// GitDivergence describes when a member's HEAD differs from the majority.
type GitDivergence struct {
	WorkspaceID  string
	MemberDID    string
	MemberHead   string
	MajorityHead string
}

// NewHealthMonitor creates a health monitor with the given configuration.
func NewHealthMonitor(cfg HealthMonitorConfig) *HealthMonitor {
	interval := cfg.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	maxMissed := cfg.MaxMissed
	if maxMissed <= 0 {
		maxMissed = 3
	}
	return &HealthMonitor{
		coordinator:  cfg.Coordinator,
		bus:          cfg.Bus,
		logger:       cfg.Logger,
		interval:     interval,
		maxMissed:    maxMissed,
		invokeFn:     cfg.InvokeFn,
		missCount:    make(map[string]map[string]int),
		lastSeen:     make(map[string]map[string]time.Time),
		gitState:     make(map[string]map[string]string),
		gitStateProv: cfg.GitStateProvider,
		workspaceIDs: cfg.WorkspaceIDsFn,
		stopCh:       make(chan struct{}),
	}
}

// Name implements lifecycle.Component.
func (h *HealthMonitor) Name() string { return "team-health-monitor" }

// Start implements lifecycle.Component. It launches the periodic health check loop
// and subscribes to task completion events for counter resets.
func (h *HealthMonitor) Start(_ context.Context, wg *sync.WaitGroup) error {
	if h.bus != nil {
		h.subsOnce.Do(func() {
			// Subscribe to task completion events to reset miss counters for successful members.
			eventbus.SubscribeTyped(h.bus, func(ev eventbus.TeamTaskCompletedEvent) {
				h.resetTeamCounters(ev.TeamID)
			})
			// Clean up maps when teams are disbanded to prevent memory leaks.
			eventbus.SubscribeTyped(h.bus, func(ev eventbus.TeamDisbandedEvent) {
				h.cleanupTeam(ev.TeamID)
			})
		})
	}

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.run()
	}()

	h.logger.Infow("health monitor started", "interval", h.interval, "maxMissed", h.maxMissed)
	return nil
}

// Stop implements lifecycle.Component.
func (h *HealthMonitor) Stop(_ context.Context) error {
	close(h.stopCh)
	h.wg.Wait()
	h.logger.Info("health monitor stopped")
	return nil
}

// run is the main health check loop.
func (h *HealthMonitor) run() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.checkAll()
		}
	}
}

// checkAll pings all active members across all active teams.
func (h *HealthMonitor) checkAll() {
	teams := h.coordinator.ActiveTeams()
	for _, t := range teams {
		if t.Status != StatusActive {
			continue
		}
		h.checkTeam(t)
	}
}

// checkTeam pings each active member of a team concurrently.
func (h *HealthMonitor) checkTeam(t *Team) {
	// Filter non-leader members once.
	allMembers := t.ActiveMembers()
	workers := make([]*Member, 0, len(allMembers))
	for _, m := range allMembers {
		if m.Role != RoleLeader {
			workers = append(workers, m)
		}
	}

	var wg sync.WaitGroup
	for _, m := range workers {
		wg.Add(1)
		go func(member *Member) {
			defer wg.Done()
			h.pingMember(t.ID, member)
		}(m)
	}
	wg.Wait()

	// Publish aggregate health check event.
	if h.bus != nil {
		healthy := 0
		for _, m := range workers {
			if h.getMissCount(t.ID, m.DID) == 0 {
				healthy++
			}
		}
		h.bus.Publish(eventbus.TeamHealthCheckEvent{
			TeamID:  t.ID,
			Healthy: healthy,
			Total:   len(workers),
		})
	}
}

// pingMember sends a health_ping to a single member and updates counters.
func (h *HealthMonitor) pingMember(teamID string, m *Member) {
	if h.invokeFn == nil {
		return
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()

	_, err := h.invokeFn(pingCtx, m.PeerID, "health_ping", map[string]interface{}{
		"teamId": teamID,
	})

	if err != nil {
		count := h.incrementMiss(teamID, m.DID)
		h.logger.Debugw("health ping failed",
			"teamID", teamID, "member", m.DID, "missCount", count, "error", err)

		if count >= h.maxMissed && h.bus != nil {
			h.bus.Publish(eventbus.TeamMemberUnhealthyEvent{
				TeamID:      teamID,
				MemberDID:   m.DID,
				MemberName:  m.Name,
				MissedPings: count,
				LastSeenAt:  h.getLastSeen(teamID, m.DID),
			})
		}
		return
	}

	// Ping succeeded: reset counter and update last seen.
	h.resetMemberCounter(teamID, m.DID)

	// Collect git state if provider is configured.
	// Use a separate context so git state calls get their own timeout budget.
	if h.gitStateProv != nil && h.workspaceIDs != nil {
		gsCtx, gsCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer gsCancel()
		for _, wsID := range h.workspaceIDs() {
			headHash, gsErr := h.gitStateProv(gsCtx, m.PeerID, wsID)
			if gsErr == nil && headHash != "" {
				h.updateGitState(wsID, m.DID, headHash)
			}
		}
	}
}

// incrementMiss increments and returns the miss counter for a member.
func (h *HealthMonitor) incrementMiss(teamID, did string) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.missCount[teamID] == nil {
		h.missCount[teamID] = make(map[string]int)
	}
	h.missCount[teamID][did]++
	return h.missCount[teamID][did]
}

// resetMemberCounter resets the miss counter and updates last seen for a member.
func (h *HealthMonitor) resetMemberCounter(teamID, did string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.missCount[teamID] != nil {
		delete(h.missCount[teamID], did)
	}
	if h.lastSeen[teamID] == nil {
		h.lastSeen[teamID] = make(map[string]time.Time)
	}
	h.lastSeen[teamID][did] = time.Now()
}

// resetTeamCounters resets all miss counters for a team (called on task completion).
func (h *HealthMonitor) resetTeamCounters(teamID string) {
	// Read team data before acquiring our own lock to avoid nested lock ordering issues.
	t, err := h.coordinator.GetTeam(teamID)
	if err != nil {
		return
	}
	members := t.ActiveMembers()

	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.missCount, teamID)
	now := time.Now()
	if h.lastSeen[teamID] == nil {
		h.lastSeen[teamID] = make(map[string]time.Time)
	}
	for _, m := range members {
		h.lastSeen[teamID][m.DID] = now
	}
}

// getMissCount returns the current miss count for a member.
func (h *HealthMonitor) getMissCount(teamID, did string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.missCount[teamID] == nil {
		return 0
	}
	return h.missCount[teamID][did]
}

// cleanupTeam removes all tracking data for a disbanded team.
func (h *HealthMonitor) cleanupTeam(teamID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.missCount, teamID)
	delete(h.lastSeen, teamID)
}

// getLastSeen returns the last time a member was seen healthy.
func (h *HealthMonitor) getLastSeen(teamID, did string) time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.lastSeen[teamID] == nil {
		return time.Time{}
	}
	return h.lastSeen[teamID][did]
}

// updateGitState records a member's HEAD hash for a workspace.
func (h *HealthMonitor) updateGitState(workspaceID, memberDID, headHash string) {
	if headHash == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.gitState[workspaceID] == nil {
		h.gitState[workspaceID] = make(map[string]string)
	}
	h.gitState[workspaceID][memberDID] = headHash
}

// DetectDivergence checks if members have different HEAD commits for a workspace.
// It returns a list of members whose HEAD differs from the majority.
func (h *HealthMonitor) DetectDivergence(workspaceID string) []GitDivergence {
	h.mu.RLock()
	defer h.mu.RUnlock()

	heads := h.gitState[workspaceID]
	if len(heads) <= 1 {
		return nil
	}

	// Count HEAD hash frequency to find majority.
	freq := make(map[string]int, len(heads))
	for _, hash := range heads {
		freq[hash]++
	}

	// Find majority HEAD.
	var majorityHead string
	maxCount := 0
	for hash, count := range freq {
		if count > maxCount {
			maxCount = count
			majorityHead = hash
		}
	}

	// Collect divergent members.
	var divergent []GitDivergence
	for did, hash := range heads {
		if hash != majorityHead {
			divergent = append(divergent, GitDivergence{
				WorkspaceID:  workspaceID,
				MemberDID:    did,
				MemberHead:   hash,
				MajorityHead: majorityHead,
			})
		}
	}

	return divergent
}
