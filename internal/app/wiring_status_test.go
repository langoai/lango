package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/types"
)

func TestStatusCollector_AddAndAll(t *testing.T) {
	t.Parallel()
	c := NewStatusCollector()

	c.Add(&types.FeatureStatus{Name: "Knowledge", Enabled: true, Healthy: true})
	c.Add(&types.FeatureStatus{Name: "Embedding", Enabled: false, Reason: "no provider"})
	c.Add(nil) // should be ignored

	all := c.All()
	assert.Len(t, all, 2)
	assert.Equal(t, "Knowledge", all[0].Name)
	assert.Equal(t, "Embedding", all[1].Name)
}

func TestStatusCollector_SilentDisabledCount(t *testing.T) {
	tests := []struct {
		give     []types.FeatureStatus
		wantCount int
	}{
		{
			give:      nil,
			wantCount: 0,
		},
		{
			give: []types.FeatureStatus{
				{Name: "Knowledge", Enabled: true, Healthy: true},
				{Name: "Embedding", Enabled: false, Reason: "no provider"},
				{Name: "Graph", Enabled: false, Reason: ""},
			},
			wantCount: 1,
		},
		{
			give: []types.FeatureStatus{
				{Name: "A", Enabled: false, Reason: "reason1"},
				{Name: "B", Enabled: false, Reason: "reason2"},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			c := NewStatusCollector()
			for i := range tt.give {
				c.Add(&tt.give[i])
			}
			assert.Equal(t, tt.wantCount, c.SilentDisabledCount())
		})
	}
}

func TestStatusCollector_AllReturnsCopy(t *testing.T) {
	t.Parallel()
	c := NewStatusCollector()
	c.Add(&types.FeatureStatus{Name: "A", Enabled: true})

	all := c.All()
	all[0].Name = "modified"

	assert.Equal(t, "A", c.All()[0].Name, "All() should return a copy")
}
