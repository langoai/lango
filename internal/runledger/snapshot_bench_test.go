package runledger

import (
	"fmt"
	"testing"
)

func BenchmarkFindStep(b *testing.B) {
	snap := &RunSnapshot{
		Steps: make([]Step, 50),
	}
	for i := range snap.Steps {
		snap.Steps[i] = Step{StepID: fmt.Sprintf("step-%d", i)}
	}
	targetID := "step-49" // worst case for linear scan
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snap.FindStep(targetID)
	}
}
