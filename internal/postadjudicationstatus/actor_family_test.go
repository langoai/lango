package postadjudicationstatus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyManualReplayActorFamily(t *testing.T) {
	tests := []struct {
		name  string
		actor string
		want  string
	}{
		{
			name:  "operator prefix",
			actor: "operator:alice",
			want:  ManualReplayActorFamilyOperator,
		},
		{
			name:  "system prefix",
			actor: "system:auto-retry",
			want:  ManualReplayActorFamilySystem,
		},
		{
			name:  "service prefix",
			actor: "service:bridge",
			want:  ManualReplayActorFamilyService,
		},
		{
			name:  "unknown empty",
			actor: "",
			want:  ManualReplayActorFamilyUnknown,
		},
		{
			name:  "unknown unmatched",
			actor: "alice",
			want:  ManualReplayActorFamilyUnknown,
		},
		{
			name:  "case insensitive",
			actor: "OPERATOR:BOB",
			want:  ManualReplayActorFamilyOperator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, ClassifyManualReplayActorFamily(tt.actor))
		})
	}
}
