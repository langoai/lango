package agent

import "testing"

func TestPIICategoryValidAndValues(t *testing.T) {
	var category PIICategory
	got := category.Values()
	want := []PIICategory{
		PIICategoryContact,
		PIICategoryIdentity,
		PIICategoryFinancial,
		PIICategoryNetwork,
	}
	if len(got) != len(want) {
		t.Fatalf("Values() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Values()[%d] = %q, want %q", i, got[i], want[i])
		}
		if !got[i].Valid() {
			t.Fatalf("%q.Valid() = false, want true", got[i])
		}
	}
	if PIICategory("unknown").Valid() {
		t.Fatal("unknown PIICategory.Valid() = true, want false")
	}
}

func TestSafetyLevelValidAndValues(t *testing.T) {
	var level SafetyLevel
	got := level.Values()
	want := []SafetyLevel{
		SafetyLevelSafe,
		SafetyLevelModerate,
		SafetyLevelDangerous,
	}
	if len(got) != len(want) {
		t.Fatalf("Values() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Values()[%d] = %v, want %v", i, got[i], want[i])
		}
		if !got[i].Valid() {
			t.Fatalf("%v.Valid() = false, want true", got[i])
		}
	}
	if SafetyLevel(99).Valid() {
		t.Fatal("unknown SafetyLevel.Valid() = true, want false")
	}
}
