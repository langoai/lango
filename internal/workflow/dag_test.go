package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDAG_Linear(t *testing.T) {
	steps := []Step{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"b"}},
	}
	dag, err := NewDAG(steps)
	require.NoError(t, err)
	require.NotNil(t, dag)

	roots := dag.Roots()
	assert.Equal(t, []string{"a"}, roots)
}

func TestNewDAG_Diamond(t *testing.T) {
	// A -> B, A -> C, B -> D, C -> D
	steps := []Step{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"a"}},
		{ID: "d", DependsOn: []string{"b", "c"}},
	}
	dag, err := NewDAG(steps)
	require.NoError(t, err)

	layers, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, layers, 3, "diamond graph should have 3 layers")
	assert.Len(t, layers[0], 1, "layer 0 should have 1 root")
	assert.Len(t, layers[1], 2, "layer 1 should have 2 nodes")
	assert.Len(t, layers[2], 1, "layer 2 should have 1 node")
}

func TestNewDAG_Parallel(t *testing.T) {
	steps := []Step{
		{ID: "a"},
		{ID: "b"},
		{ID: "c"},
	}
	dag, err := NewDAG(steps)
	require.NoError(t, err)

	layers, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, layers, 1, "all-parallel graph should have 1 layer")
	assert.Len(t, layers[0], 3)
}

func TestNewDAG_CircularDependency(t *testing.T) {
	steps := []Step{
		{ID: "a", DependsOn: []string{"b"}},
		{ID: "b", DependsOn: []string{"a"}},
	}
	dag, err := NewDAG(steps)
	assert.Error(t, err)
	assert.Nil(t, dag)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestTopologicalSort_Layers(t *testing.T) {
	// a -> b -> c (linear chain)
	steps := []Step{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"b"}},
	}
	dag, err := NewDAG(steps)
	require.NoError(t, err)

	layers, err := dag.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, layers, 3)
	assert.Contains(t, layers[0], "a")
	assert.Contains(t, layers[1], "b")
	assert.Contains(t, layers[2], "c")
}

func TestRoots(t *testing.T) {
	steps := []Step{
		{ID: "root1"},
		{ID: "root2"},
		{ID: "child", DependsOn: []string{"root1", "root2"}},
	}
	dag, err := NewDAG(steps)
	require.NoError(t, err)

	roots := dag.Roots()
	assert.Len(t, roots, 2)
	assert.Contains(t, roots, "root1")
	assert.Contains(t, roots, "root2")
}

func TestReady_NoneCompleted(t *testing.T) {
	steps := []Step{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
	}
	dag, err := NewDAG(steps)
	require.NoError(t, err)

	ready := dag.Ready(map[string]bool{})
	assert.Equal(t, []string{"a"}, ready, "only root should be ready when nothing is completed")
}

func TestReady_SomeCompleted(t *testing.T) {
	steps := []Step{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"a"}},
	}
	dag, err := NewDAG(steps)
	require.NoError(t, err)

	ready := dag.Ready(map[string]bool{"a": true})
	assert.Len(t, ready, 2)
	assert.Contains(t, ready, "b")
	assert.Contains(t, ready, "c")
}

func TestReady_AllCompleted(t *testing.T) {
	steps := []Step{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
	}
	dag, err := NewDAG(steps)
	require.NoError(t, err)

	ready := dag.Ready(map[string]bool{"a": true, "b": true})
	assert.Empty(t, ready)
}
