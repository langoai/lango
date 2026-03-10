package team

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/agentpool"
)

// Sentinel errors for coordination.
var (
	ErrTeamNotFound    = errors.New("team not found")
	ErrInsufficientAck = errors.New("insufficient acknowledgments from team members")
)

// ConflictStrategy defines how to resolve conflicting results from multiple agents.
type ConflictStrategy string

const (
	StrategyTrustWeighted  ConflictStrategy = "trust_weighted"
	StrategyMajorityVote   ConflictStrategy = "majority_vote"
	StrategyLeaderDecides  ConflictStrategy = "leader_decides"
	StrategyFailOnConflict ConflictStrategy = "fail_on_conflict"
)

// AssignmentStrategy determines how tasks are assigned to members.
type AssignmentStrategy string

const (
	AssignBestMatch    AssignmentStrategy = "best_match"
	AssignRoundRobin   AssignmentStrategy = "round_robin"
	AssignLoadBalanced AssignmentStrategy = "load_balanced"
)

// TaskResult holds the output of a delegated task from a single member.
type TaskResult struct {
	MemberDID string
	Result    map[string]interface{}
	Err       error
	Duration  time.Duration
}

// InvokeFunc is the callback used by the coordinator to send a task to a remote agent.
type InvokeFunc func(ctx context.Context, peerID, toolName string, params map[string]interface{}) (map[string]interface{}, error)

// ConflictResolver decides the final result when members produce conflicting outputs.
type ConflictResolver func(results []TaskResult) (map[string]interface{}, error)

// MajorityResolver picks the most common result by simple majority.
func MajorityResolver(results []TaskResult) (map[string]interface{}, error) {
	if len(results) == 0 {
		return nil, ErrConflict
	}

	// Count successful results.
	var successful []TaskResult
	for _, r := range results {
		if r.Err == nil && r.Result != nil {
			successful = append(successful, r)
		}
	}
	if len(successful) == 0 {
		return nil, fmt.Errorf("all team members failed: %w", ErrConflict)
	}

	// For simplicity, return the first successful result.
	// A production implementation would hash results and pick the majority.
	return successful[0].Result, nil
}

// FastestResolver picks the first successful result.
func FastestResolver(results []TaskResult) (map[string]interface{}, error) {
	for _, r := range results {
		if r.Err == nil && r.Result != nil {
			return r.Result, nil
		}
	}
	return nil, fmt.Errorf("all team members failed: %w", ErrConflict)
}

// CoordinatorConfig configures the team coordinator.
type CoordinatorConfig struct {
	Pool             *agentpool.Pool
	Selector         *agentpool.Selector
	InvokeFn         InvokeFunc
	ConflictResolver ConflictResolver
	Conflict         ConflictStrategy
	Assignment       AssignmentStrategy
	Bus              *eventbus.Bus
	Logger           *zap.SugaredLogger
}

// Coordinator manages the lifecycle of agent teams — forming, delegating, collecting, and disbanding.
type Coordinator struct {
	pool       *agentpool.Pool
	selector   *agentpool.Selector
	invokeFn   InvokeFunc
	resolver   ConflictResolver
	conflict   ConflictStrategy
	assignment AssignmentStrategy
	bus        *eventbus.Bus
	logger     *zap.SugaredLogger

	mu    sync.RWMutex
	teams map[string]*Team
}

// NewCoordinator creates a team coordinator.
func NewCoordinator(cfg CoordinatorConfig) *Coordinator {
	resolver := cfg.ConflictResolver
	if resolver == nil {
		resolver = MajorityResolver
	}
	conflict := cfg.Conflict
	if conflict == "" {
		conflict = StrategyMajorityVote
	}
	assignment := cfg.Assignment
	if assignment == "" {
		assignment = AssignBestMatch
	}
	return &Coordinator{
		pool:       cfg.Pool,
		selector:   cfg.Selector,
		invokeFn:   cfg.InvokeFn,
		resolver:   resolver,
		conflict:   conflict,
		assignment: assignment,
		bus:        cfg.Bus,
		logger:     cfg.Logger,
		teams:      make(map[string]*Team),
	}
}

// FormTeamRequest describes how to form a new team.
type FormTeamRequest struct {
	TeamID      string
	Name        string
	Goal        string
	LeaderDID   string
	Capability  string
	MemberCount int
	MaxMembers  int
}

// FormTeam creates a new team by selecting agents from the pool.
func (c *Coordinator) FormTeam(ctx context.Context, req FormTeamRequest) (*Team, error) {
	maxMembers := req.MaxMembers
	if maxMembers <= 0 {
		maxMembers = req.MemberCount + 1 // leader + workers
	}

	t := NewTeam(req.TeamID, req.Name, req.Goal, req.LeaderDID, maxMembers)

	// Add leader.
	leader := c.pool.Get(req.LeaderDID)
	if leader != nil {
		if err := t.AddMember(&Member{
			DID:          leader.DID,
			Name:         leader.Name,
			PeerID:       leader.PeerID,
			Role:         RoleLeader,
			Capabilities: leader.Capabilities,
		}); err != nil {
			return nil, fmt.Errorf("add leader: %w", err)
		}
	}

	// Select workers.
	if req.MemberCount > 0 && req.Capability != "" {
		agents, err := c.selector.SelectN(req.Capability, req.MemberCount)
		if err != nil {
			return nil, fmt.Errorf("select agents: %w", err)
		}

		for _, a := range agents {
			if a.DID == req.LeaderDID {
				continue
			}
			if err := t.AddMember(&Member{
				DID:          a.DID,
				Name:         a.Name,
				PeerID:       a.PeerID,
				Role:         RoleWorker,
				Capabilities: a.Capabilities,
			}); err != nil {
				c.logger.Debugw("skip member during formation", "did", a.DID, "error", err)
			}
		}
	}

	t.Activate()

	c.mu.Lock()
	c.teams[t.ID] = t
	c.mu.Unlock()

	c.logger.Infow("team formed",
		"teamID", t.ID,
		"name", t.Name,
		"members", t.MemberCount(),
	)

	// Publish team-formed event.
	if c.bus != nil {
		c.bus.Publish(eventbus.TeamFormedEvent{
			TeamID:    t.ID,
			Name:      t.Name,
			Goal:      t.Goal,
			LeaderDID: t.LeaderDID,
			Members:   t.MemberCount(),
		})
	}

	// Publish events for each member that joined.
	if c.bus != nil {
		for _, m := range t.Members() {
			c.bus.Publish(eventbus.TeamMemberJoinedEvent{
				TeamID:    t.ID,
				MemberDID: m.DID,
				Role:      string(m.Role),
			})
		}
	}

	return t, nil
}

// GetTeam returns a team by ID.
func (c *Coordinator) GetTeam(teamID string) (*Team, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok := c.teams[teamID]
	if !ok {
		return nil, ErrTeamNotFound
	}
	return t, nil
}

// DelegateTask sends a task to all workers in the team and collects results.
func (c *Coordinator) DelegateTask(ctx context.Context, teamID, toolName string, params map[string]interface{}) ([]TaskResult, error) {
	t, err := c.GetTeam(teamID)
	if err != nil {
		return nil, err
	}

	members := t.Members()
	var workers []*Member
	for _, m := range members {
		if m.Role == RoleWorker {
			workers = append(workers, m)
		}
	}

	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers in team %q", teamID)
	}

	// Publish task-delegated event.
	if c.bus != nil {
		c.bus.Publish(eventbus.TeamTaskDelegatedEvent{
			TeamID:   teamID,
			ToolName: toolName,
			Workers:  len(workers),
		})
	}

	// Dispatch to all workers concurrently.
	results := make([]TaskResult, len(workers))
	var wg sync.WaitGroup

	for i, w := range workers {
		wg.Add(1)
		go func(idx int, member *Member) {
			defer wg.Done()

			scopedCtx := WithScopedContext(ctx, ScopedContext{
				TeamID:    teamID,
				MemberDID: member.DID,
				Role:      member.Role,
			})

			start := time.Now()
			result, invokeErr := c.invokeFn(scopedCtx, member.PeerID, toolName, params)
			results[idx] = TaskResult{
				MemberDID: member.DID,
				Result:    result,
				Err:       invokeErr,
				Duration:  time.Since(start),
			}
		}(i, w)
	}

	wg.Wait()

	// Publish task-completed event.
	if c.bus != nil {
		var successful, failed int
		var totalDuration time.Duration
		for _, r := range results {
			if r.Err == nil {
				successful++
			} else {
				failed++
			}
			totalDuration += r.Duration
		}
		c.bus.Publish(eventbus.TeamTaskCompletedEvent{
			TeamID:     teamID,
			ToolName:   toolName,
			Successful: successful,
			Failed:     failed,
			Duration:   totalDuration / time.Duration(len(results)),
		})
	}

	return results, nil
}

// CollectResults resolves conflicts from delegated task results using the configured resolver.
func (c *Coordinator) CollectResults(teamID, toolName string, results []TaskResult) (map[string]interface{}, error) {
	resolved, err := c.resolver(results)
	if err != nil && c.bus != nil {
		// Count unique successful members for conflict detection.
		var successCount int
		for _, r := range results {
			if r.Err == nil {
				successCount++
			}
		}
		if successCount > 1 {
			c.bus.Publish(eventbus.TeamConflictDetectedEvent{
				TeamID:   teamID,
				ToolName: toolName,
				Members:  successCount,
			})
		}
	}
	return resolved, err
}

// DisbandTeam marks a team as disbanded and removes it from the coordinator.
func (c *Coordinator) DisbandTeam(teamID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	t, ok := c.teams[teamID]
	if !ok {
		return ErrTeamNotFound
	}

	// Publish leave events before disbanding.
	if c.bus != nil {
		for _, m := range t.Members() {
			c.bus.Publish(eventbus.TeamMemberLeftEvent{
				TeamID:    teamID,
				MemberDID: m.DID,
				Reason:    "team disbanded",
			})
		}
	}

	t.Disband()
	delete(c.teams, teamID)

	if c.bus != nil {
		c.bus.Publish(eventbus.TeamDisbandedEvent{
			TeamID: teamID,
			Reason: "team disbanded",
		})
	}

	c.logger.Infow("team disbanded", "teamID", teamID, "name", t.Name)
	return nil
}

// ActiveTeams returns all currently managed teams (alias for ListTeams).
func (c *Coordinator) ActiveTeams() []*Team {
	return c.ListTeams()
}

// ListTeams returns all active teams.
func (c *Coordinator) ListTeams() []*Team {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*Team, 0, len(c.teams))
	for _, t := range c.teams {
		result = append(result, t)
	}
	return result
}
