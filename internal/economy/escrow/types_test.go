package escrow

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompletedMilestones(t *testing.T) {
	tests := []struct {
		give       string
		milestones []Milestone
		want       int
	}{
		{
			give:       "no milestones",
			milestones: nil,
			want:       0,
		},
		{
			give: "all pending",
			milestones: []Milestone{
				{ID: "m1", Status: MilestonePending},
				{ID: "m2", Status: MilestonePending},
			},
			want: 0,
		},
		{
			give: "one completed",
			milestones: []Milestone{
				{ID: "m1", Status: MilestoneCompleted},
				{ID: "m2", Status: MilestonePending},
			},
			want: 1,
		},
		{
			give: "all completed",
			milestones: []Milestone{
				{ID: "m1", Status: MilestoneCompleted},
				{ID: "m2", Status: MilestoneCompleted},
			},
			want: 2,
		},
		{
			give: "mixed statuses",
			milestones: []Milestone{
				{ID: "m1", Status: MilestoneCompleted},
				{ID: "m2", Status: MilestoneDisputed},
				{ID: "m3", Status: MilestonePending},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			entry := &EscrowEntry{Milestones: tt.milestones}
			assert.Equal(t, tt.want, entry.CompletedMilestones())
		})
	}
}

func TestAllMilestonesCompleted(t *testing.T) {
	now := time.Now()
	tests := []struct {
		give       string
		milestones []Milestone
		want       bool
	}{
		{
			give:       "no milestones returns false",
			milestones: nil,
			want:       false,
		},
		{
			give: "not all completed",
			milestones: []Milestone{
				{ID: "m1", Status: MilestoneCompleted, CompletedAt: &now},
				{ID: "m2", Status: MilestonePending},
			},
			want: false,
		},
		{
			give: "all completed",
			milestones: []Milestone{
				{ID: "m1", Status: MilestoneCompleted, CompletedAt: &now},
				{ID: "m2", Status: MilestoneCompleted, CompletedAt: &now},
			},
			want: true,
		},
		{
			give: "single completed",
			milestones: []Milestone{
				{ID: "m1", Status: MilestoneCompleted, CompletedAt: &now},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			entry := &EscrowEntry{
				Milestones:  tt.milestones,
				TotalAmount: big.NewInt(100),
			}
			assert.Equal(t, tt.want, entry.AllMilestonesCompleted())
		})
	}
}
