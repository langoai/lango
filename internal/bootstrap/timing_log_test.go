package bootstrap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTimingLogDir(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, timingLogFile)

	origFn := timingLogPathFn
	timingLogPathFn = func() (string, error) { return path, nil }
	return path, func() { timingLogPathFn = origFn }
}

func sampleEntries(n int) []PhaseTimingEntry {
	return []PhaseTimingEntry{
		{Phase: "config", Duration: time.Duration(n) * time.Millisecond},
		{Phase: "db", Duration: time.Duration(n*2) * time.Millisecond},
	}
}

func TestAppendTimingLog_Write(t *testing.T) {
	path, cleanup := setupTimingLogDir(t)
	defer cleanup()

	err := AppendTimingLog(sampleEntries(10), "v1.0.0")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var entry TimingLogEntry
	require.NoError(t, json.Unmarshal(data, &entry))
	assert.Equal(t, "v1.0.0", entry.Version)
	assert.Len(t, entry.Phases, 2)
	assert.Equal(t, "config", entry.Phases[0].Name)
	assert.Equal(t, int64(10), entry.Phases[0].DurationMs)
}

func TestAppendTimingLog_Rotation(t *testing.T) {
	_, cleanup := setupTimingLogDir(t)
	defer cleanup()

	for i := 0; i < maxTimingEntries+5; i++ {
		err := AppendTimingLog(sampleEntries(i), "v1.0.0")
		require.NoError(t, err)
	}

	entries, err := ReadTimingLog()
	require.NoError(t, err)
	assert.Len(t, entries, maxTimingEntries)

	var first TimingLogEntry
	first = entries[0]
	assert.Equal(t, int64(5), first.Phases[0].DurationMs)
}

func TestReadTimingLog_CorruptedLines(t *testing.T) {
	path, cleanup := setupTimingLogDir(t)
	defer cleanup()

	require.NoError(t, AppendTimingLog(sampleEntries(10), "v1"))

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, timingFilePerm)
	require.NoError(t, err)
	_, _ = f.WriteString("this is garbage\n")
	f.Close()

	require.NoError(t, AppendTimingLog(sampleEntries(20), "v2"))

	entries, err := ReadTimingLog()
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "v1", entries[0].Version)
	assert.Equal(t, "v2", entries[1].Version)
}

func TestAppendTimingLog_MissingFile(t *testing.T) {
	_, cleanup := setupTimingLogDir(t)
	defer cleanup()

	entries, err := ReadTimingLog()
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestAppendTimingLog_WriteFail(t *testing.T) {
	origFn := timingLogPathFn
	timingLogPathFn = func() (string, error) { return "/nonexistent/path/timing.jsonl", nil }
	defer func() { timingLogPathFn = origFn }()

	err := AppendTimingLog(sampleEntries(10), "v1")
	assert.Error(t, err)
}
