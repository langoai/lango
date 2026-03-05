package agentpool

// DynamicAgentInfo describes a discovered agent for routing purposes.
// It is a lightweight descriptor that avoids importing ADK agent types.
type DynamicAgentInfo struct {
	Name         string
	DID          string
	PeerID       string
	Description  string
	Capabilities []string
	TrustScore   float64
	PricePerCall float64
}

// DynamicAgentProvider discovers remote agents dynamically at runtime.
// The orchestrator queries this interface to integrate P2P agents into
// its routing table without requiring them to implement adk_agent.Agent.
type DynamicAgentProvider interface {
	// AvailableAgents returns all healthy agents currently in the pool.
	AvailableAgents() []DynamicAgentInfo

	// FindForCapability returns agents that match the given capability.
	FindForCapability(capability string) []DynamicAgentInfo
}

// Compile-time interface check.
var _ DynamicAgentProvider = (*PoolProvider)(nil)

// PoolProvider adapts an agentpool.Pool into a DynamicAgentProvider.
type PoolProvider struct {
	pool     *Pool
	selector *Selector
}

// NewPoolProvider creates a DynamicAgentProvider backed by a Pool.
func NewPoolProvider(pool *Pool, selector *Selector) *PoolProvider {
	return &PoolProvider{pool: pool, selector: selector}
}

// AvailableAgents returns all healthy agents in the pool.
func (p *PoolProvider) AvailableAgents() []DynamicAgentInfo {
	agents := p.pool.List()
	result := make([]DynamicAgentInfo, 0, len(agents))
	for _, a := range agents {
		if a.Status == StatusUnhealthy {
			continue
		}
		result = append(result, agentToInfo(a))
	}
	return result
}

// FindForCapability returns agents matching the given capability.
func (p *PoolProvider) FindForCapability(capability string) []DynamicAgentInfo {
	agents := p.pool.FindByCapability(capability)
	result := make([]DynamicAgentInfo, 0, len(agents))
	for _, a := range agents {
		result = append(result, agentToInfo(a))
	}
	return result
}

func agentToInfo(a *Agent) DynamicAgentInfo {
	desc := a.Name
	if len(a.Capabilities) > 0 {
		desc = a.Name + " (P2P remote agent)"
	}
	return DynamicAgentInfo{
		Name:         a.Name,
		DID:          a.DID,
		PeerID:       a.PeerID,
		Description:  desc,
		Capabilities: a.Capabilities,
		TrustScore:   a.TrustScore,
		PricePerCall: a.PricePerCall,
	}
}
