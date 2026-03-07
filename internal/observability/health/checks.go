package health

import (
	"context"
	"runtime"
	"strconv"
	"time"
)

// DatabaseCheck checks database connectivity.
type DatabaseCheck struct {
	ping func(ctx context.Context) error
}

// NewDatabaseCheck creates a new DatabaseCheck.
// ping should be a function that tests DB connectivity (e.g., db.PingContext).
func NewDatabaseCheck(ping func(ctx context.Context) error) *DatabaseCheck {
	return &DatabaseCheck{ping: ping}
}

func (c *DatabaseCheck) Name() string { return "database" }

func (c *DatabaseCheck) Check(ctx context.Context) ComponentHealth {
	ch := ComponentHealth{
		Name:        c.Name(),
		LastChecked: time.Now(),
	}
	if err := c.ping(ctx); err != nil {
		ch.Status = StatusUnhealthy
		ch.Message = err.Error()
		return ch
	}
	ch.Status = StatusHealthy
	ch.Message = "connected"
	return ch
}

// MemoryCheck reports runtime memory stats.
type MemoryCheck struct {
	threshold uint64 // bytes; above this -> degraded
}

// NewMemoryCheck creates a new MemoryCheck with the given threshold in bytes.
func NewMemoryCheck(thresholdBytes uint64) *MemoryCheck {
	return &MemoryCheck{threshold: thresholdBytes}
}

func (c *MemoryCheck) Name() string { return "memory" }

func (c *MemoryCheck) Check(_ context.Context) ComponentHealth {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	ch := ComponentHealth{
		Name:        c.Name(),
		Status:      StatusHealthy,
		LastChecked: time.Now(),
		Metadata: map[string]string{
			"allocMB":    strconv.FormatUint(ms.Alloc/(1024*1024), 10),
			"sysMB":      strconv.FormatUint(ms.Sys/(1024*1024), 10),
			"goroutines": strconv.Itoa(runtime.NumGoroutine()),
		},
	}
	if c.threshold > 0 && ms.Alloc > c.threshold {
		ch.Status = StatusDegraded
		ch.Message = "memory usage above threshold"
	}
	return ch
}

// ProviderCheck verifies that an LLM provider is reachable.
type ProviderCheck struct {
	name string
	ping func(ctx context.Context) error
}

// NewProviderCheck creates a new ProviderCheck.
func NewProviderCheck(name string, ping func(ctx context.Context) error) *ProviderCheck {
	return &ProviderCheck{name: name, ping: ping}
}

func (c *ProviderCheck) Name() string { return "provider." + c.name }

func (c *ProviderCheck) Check(ctx context.Context) ComponentHealth {
	ch := ComponentHealth{
		Name:        c.Name(),
		LastChecked: time.Now(),
	}
	if err := c.ping(ctx); err != nil {
		ch.Status = StatusDegraded
		ch.Message = err.Error()
		return ch
	}
	ch.Status = StatusHealthy
	ch.Message = "reachable"
	return ch
}
