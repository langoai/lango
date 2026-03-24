package turntrace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRetentionCleaner_Defaults(t *testing.T) {
	c := NewRetentionCleaner(nil, RetentionConfig{})
	assert.Equal(t, time.Hour, c.config.CleanupInterval)
	assert.Equal(t, 30*24*time.Hour, c.config.MaxAge)
	assert.Equal(t, 10000, c.config.MaxTraces)
	assert.Equal(t, 2, c.config.FailedTraceMultiplier)
}

func TestNewRetentionCleaner_CustomConfig(t *testing.T) {
	cfg := RetentionConfig{
		MaxAge:                48 * time.Hour,
		MaxTraces:             500,
		FailedTraceMultiplier: 3,
		CleanupInterval:       10 * time.Minute,
	}
	c := NewRetentionCleaner(nil, cfg)
	assert.Equal(t, 48*time.Hour, c.config.MaxAge)
	assert.Equal(t, 500, c.config.MaxTraces)
	assert.Equal(t, 3, c.config.FailedTraceMultiplier)
	assert.Equal(t, 10*time.Minute, c.config.CleanupInterval)
}

func TestRetentionCleaner_Name(t *testing.T) {
	c := NewRetentionCleaner(nil, RetentionConfig{})
	assert.Equal(t, "turntrace-retention", c.Name())
}

func TestRetentionCleaner_StartStop(t *testing.T) {
	c := NewRetentionCleaner(nil, RetentionConfig{CleanupInterval: time.Hour})
	err := c.Start(nil, nil)
	assert.NoError(t, err)
	err = c.Stop(nil)
	assert.NoError(t, err)
}
