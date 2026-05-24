package postadjudicationstatus

import "testing"

func TestClassifyDispatchReferenceFamily(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "hyphenated dispatch", input: "dispatch-final-123", want: DispatchReferenceFamilyDispatch},
		{name: "slash separated queue", input: "queue/retry/123", want: DispatchReferenceFamilyQueue},
		{name: "colon separated bridge", input: "bridge:dead-letter:1", want: DispatchReferenceFamilyBridge},
		{name: "underscore worker", input: "worker_retry_77", want: DispatchReferenceFamilyWorker},
		{name: "single token", input: "webhook", want: DispatchReferenceFamilyWebhook},
		{name: "empty", input: "   ", want: DispatchReferenceFamilyUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyDispatchReferenceFamily(tt.input); got != tt.want {
				t.Fatalf("ClassifyDispatchReferenceFamily(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
