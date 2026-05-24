package postadjudicationstatus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyDeadLetterReasonFamily(t *testing.T) {
	tests := []struct {
		name   string
		reason string
		want   string
	}{
		{
			name:   "retry exhausted",
			reason: "retry attempts exhausted after 5 attempts",
			want:   DeadLetterReasonFamilyRetryExhausted,
		},
		{
			name:   "policy blocked",
			reason: "policy gate denied replay",
			want:   DeadLetterReasonFamilyPolicyBlocked,
		},
		{
			name:   "invalid receipt",
			reason: "invalid transaction receipt evidence",
			want:   DeadLetterReasonFamilyReceiptInvalid,
		},
		{
			name:   "invalid adjudication",
			reason: "adjudication invalid for transaction evidence",
			want:   DeadLetterReasonFamilyReceiptInvalid,
		},
		{
			name:   "background failed",
			reason: "background dispatch worker failed",
			want:   DeadLetterReasonFamilyBackgroundFailed,
		},
		{
			name:   "unknown empty",
			reason: "",
			want:   DeadLetterReasonFamilyUnknown,
		},
		{
			name:   "unknown unmatched",
			reason: "unexpected storage condition",
			want:   DeadLetterReasonFamilyUnknown,
		},
		{
			name:   "case insensitive",
			reason: "POLICY BLOCKED BY GATE",
			want:   DeadLetterReasonFamilyPolicyBlocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, ClassifyDeadLetterReasonFamily(tt.reason))
		})
	}
}
