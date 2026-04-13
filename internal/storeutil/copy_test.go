package storeutil

import "testing"

func TestCopySlice(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var src []string
		got := CopySlice(src)
		if got != nil {
			t.Errorf("CopySlice(nil) should be nil")
		}
	})
	t.Run("independent", func(t *testing.T) {
		src := []string{"a", "b", "c"}
		got := CopySlice(src)
		got[0] = "X"
		if src[0] != "a" {
			t.Errorf("modifying copy affected original")
		}
	})
}

func TestCopyMap(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var src map[string]int
		got := CopyMap(src)
		if got != nil {
			t.Errorf("CopyMap(nil) should be nil")
		}
	})
	t.Run("independent", func(t *testing.T) {
		src := map[string]int{"a": 1, "b": 2}
		got := CopyMap(src)
		got["a"] = 99
		if src["a"] != 1 {
			t.Errorf("modifying copy affected original")
		}
	})
}
