package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_MetadataAndReadStatHandlers(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notes.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("one\ntwo\nthree"), 0644))

	tools := BuildTools(New(Config{AllowedPaths: []string{tmpDir}}))
	byName := buildToolsMetadataAndReadStatHandlersFilesystemToolsByName(tools)

	require.Len(t, tools, 7)
	assert.ElementsMatch(t, []string{
		"fs_read",
		"fs_list",
		"fs_write",
		"fs_edit",
		"fs_mkdir",
		"fs_delete",
		"fs_stat",
	}, buildToolsMetadataAndReadStatHandlersFilesystemToolNames(tools))

	readTool := byName["fs_read"]
	require.NotNil(t, readTool)
	assert.Equal(t, agent.SafetyLevelSafe, readTool.SafetyLevel)
	assert.Equal(t, agent.ActivityRead, readTool.Capability.Activity)
	assert.True(t, readTool.Capability.ReadOnly)
	assert.True(t, readTool.Capability.ConcurrencySafe)
	assert.Contains(t, readTool.Capability.Aliases, "cat")

	readResult, err := readTool.Handler(context.Background(), map[string]interface{}{
		"path":   filePath,
		"offset": float64(2),
		"limit":  float64(1),
	})
	require.NoError(t, err)
	readMeta, ok := readResult.(*ReadResult)
	require.True(t, ok)
	assert.Equal(t, "two", readMeta.Content)
	assert.Equal(t, 3, readMeta.TotalLines)
	assert.Equal(t, 2, readMeta.Offset)
	assert.Equal(t, 1, readMeta.Limit)

	statTool := byName["fs_stat"]
	require.NotNil(t, statTool)
	assert.Equal(t, agent.ActivityQuery, statTool.Capability.Activity)
	assert.True(t, statTool.Capability.ReadOnly)
	assert.Contains(t, statTool.Capability.Aliases, "file_info")

	statResult, err := statTool.Handler(context.Background(), map[string]interface{}{"path": filePath})
	require.NoError(t, err)
	stat, ok := statResult.(*StatResult)
	require.True(t, ok)
	wantStatPath, err := filepath.EvalSymlinks(filePath)
	require.NoError(t, err)
	assert.Equal(t, wantStatPath, stat.Path)
	assert.Equal(t, 3, stat.Lines)
	assert.False(t, stat.IsDir)
}

func TestBuildTools_MutationHandlersRequireInputsAndApplySideEffects(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tools := buildToolsMetadataAndReadStatHandlersFilesystemToolsByName(BuildTools(New(Config{AllowedPaths: []string{tmpDir}})))

	mkdirResult, err := tools["fs_mkdir"].Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, mkdirResult)
	assert.ErrorContains(t, err, "missing path parameter")

	dirPath := filepath.Join(tmpDir, "created")
	mkdirResult, err = tools["fs_mkdir"].Handler(context.Background(), map[string]interface{}{"path": dirPath})
	require.NoError(t, err)
	assert.Nil(t, mkdirResult)
	assert.DirExists(t, dirPath)

	filePath := filepath.Join(tmpDir, "delete-me.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("delete me"), 0644))
	deleteResult, err := tools["fs_delete"].Handler(context.Background(), map[string]interface{}{"path": filePath})
	require.NoError(t, err)
	assert.Nil(t, deleteResult)
	assert.NoFileExists(t, filePath)
}

func buildToolsMetadataAndReadStatHandlersFilesystemToolsByName(tools []*agent.Tool) map[string]*agent.Tool {
	byName := make(map[string]*agent.Tool, len(tools))
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	return byName
}

func buildToolsMetadataAndReadStatHandlersFilesystemToolNames(tools []*agent.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	return names
}
