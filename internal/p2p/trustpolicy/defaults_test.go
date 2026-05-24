package trustpolicy

import "testing"

func TestDefaultPostPayThreshold(t *testing.T) {
	t.Parallel()

	if DefaultPostPayThreshold != 0.8 {
		t.Fatalf("DefaultPostPayThreshold = %v, want 0.8", DefaultPostPayThreshold)
	}
}
