package health

import (
	"context"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// ComponentHealth is the result of a single health check.
type ComponentHealth struct {
	Name        string            `json:"name"`
	Status      Status            `json:"status"`
	Message     string            `json:"message,omitempty"`
	LastChecked time.Time         `json:"lastChecked"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SystemHealth aggregates all component health checks.
type SystemHealth struct {
	Status     Status            `json:"status"`
	Components []ComponentHealth `json:"components"`
	CheckedAt  time.Time         `json:"checkedAt"`
}

// Checker is the interface for health checks.
type Checker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}
