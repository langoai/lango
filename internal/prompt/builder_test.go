package prompt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Add_And_Build(t *testing.T) {
	b := NewBuilder()
	b.Add(NewStaticSection("b", 200, "B", "second"))
	b.Add(NewStaticSection("a", 100, "A", "first"))

	result := b.Build()
	idxA := strings.Index(result, "first")
	idxB := strings.Index(result, "second")
	require.NotEqual(t, -1, idxA)
	require.NotEqual(t, -1, idxB)
	assert.Less(t, idxA, idxB, "lower priority section should appear first")
}

func TestBuilder_Add_ReplacesExistingID(t *testing.T) {
	b := NewBuilder()
	b.Add(NewStaticSection("id1", 100, "Old", "old content"))
	b.Add(NewStaticSection("id1", 100, "New", "new content"))

	result := b.Build()
	assert.Contains(t, result, "new content")
	assert.NotContains(t, result, "old content")
}

func TestBuilder_Remove(t *testing.T) {
	b := NewBuilder()
	b.Add(NewStaticSection("keep", 100, "Keep", "keep me"))
	b.Add(NewStaticSection("drop", 200, "Drop", "drop me"))
	b.Remove("drop")

	result := b.Build()
	assert.Contains(t, result, "keep me")
	assert.NotContains(t, result, "drop me")
}

func TestBuilder_Remove_NonExistent(t *testing.T) {
	b := NewBuilder()
	b.Add(NewStaticSection("a", 100, "A", "content"))
	b.Remove("nonexistent") // should not panic
	assert.True(t, b.Has("a"))
}

func TestBuilder_Has(t *testing.T) {
	b := NewBuilder()
	assert.False(t, b.Has("missing"))

	b.Add(NewStaticSection("present", 100, "", "content"))
	assert.True(t, b.Has("present"))
}

func TestBuilder_Build_SkipsEmptyRenders(t *testing.T) {
	b := NewBuilder()
	b.Add(NewStaticSection("visible", 100, "", "content"))
	b.Add(NewStaticSection("empty", 200, "Empty", ""))

	result := b.Build()
	assert.Equal(t, "content", result)
}

func TestBuilder_Build_Empty(t *testing.T) {
	b := NewBuilder()
	assert.Equal(t, "", b.Build())
}

func TestBuilder_Clone(t *testing.T) {
	b := NewBuilder()
	b.Add(NewStaticSection("a", 100, "A", "alpha"))
	b.Add(NewStaticSection("b", 200, "B", "bravo"))

	clone := b.Clone()

	// Mutate clone — original must be unaffected.
	clone.Add(NewStaticSection("c", 300, "C", "charlie"))
	clone.Add(NewStaticSection("a", 100, "A", "alpha-override"))

	original := b.Build()
	assert.Contains(t, original, "alpha")
	assert.NotContains(t, original, "charlie")
	assert.NotContains(t, original, "alpha-override")

	cloned := clone.Build()
	assert.Contains(t, cloned, "alpha-override")
	assert.Contains(t, cloned, "charlie")
}

func TestBuilder_Clone_Empty(t *testing.T) {
	b := NewBuilder()
	clone := b.Clone()
	clone.Add(NewStaticSection("x", 100, "", "x"))
	assert.Equal(t, "", b.Build())
	assert.Equal(t, "x", clone.Build())
}

func TestBuilder_PrioritySorting(t *testing.T) {
	b := NewBuilder()
	b.Add(NewStaticSection("c", 300, "", "third"))
	b.Add(NewStaticSection("a", 100, "", "first"))
	b.Add(NewStaticSection("b", 200, "", "second"))

	result := b.Build()
	parts := strings.Split(result, "\n\n")
	require.Len(t, parts, 3)
	assert.Equal(t, "first", parts[0])
	assert.Equal(t, "second", parts[1])
	assert.Equal(t, "third", parts[2])
}
