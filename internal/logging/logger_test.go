package logging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

// NOTE: Tests that call Init() or access package-level globals (rootLogger, sugarLogger)
// cannot be run in parallel because Init writes to shared globals without synchronization.

func TestInit_JSONFormat(t *testing.T) {
	err := Init(LogConfig{
		Level:  "debug",
		Format: "json",
	})
	require.NoError(t, err)
}

func TestInit_ConsoleFormat(t *testing.T) {
	err := Init(LogConfig{
		Level:  "info",
		Format: "console",
	})
	require.NoError(t, err)
}

func TestInit_OutputToFile(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	err := Init(LogConfig{
		Level:      "warn",
		Format:     "json",
		OutputPath: logFile,
	})
	require.NoError(t, err)

	_, statErr := os.Stat(logFile)
	assert.NoError(t, statErr)
}

func TestInit_InvalidOutputPath(t *testing.T) {
	err := Init(LogConfig{
		Level:      "info",
		OutputPath: "/nonexistent/dir/test.log",
	})
	require.Error(t, err)
}

func TestInit_UnknownLevel_DefaultsToInfo(t *testing.T) {
	err := Init(LogConfig{
		Level: "unknown_level",
	})
	require.NoError(t, err)
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		wantLevel zapcore.Level
	}{
		{give: "debug", wantLevel: zapcore.DebugLevel},
		{give: "info", wantLevel: zapcore.InfoLevel},
		{give: "warn", wantLevel: zapcore.WarnLevel},
		{give: "error", wantLevel: zapcore.ErrorLevel},
		{give: "unknown", wantLevel: zapcore.InfoLevel},
		{give: "", wantLevel: zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			level, _ := parseLevel(tt.give)
			assert.Equal(t, tt.wantLevel, level)
		})
	}
}

func TestLogger_NotInitialized(t *testing.T) {
	origRoot := rootLogger
	origSugar := sugarLogger
	rootLogger = nil
	sugarLogger = nil
	t.Cleanup(func() {
		rootLogger = origRoot
		sugarLogger = origSugar
	})

	logger := Logger()
	require.NotNil(t, logger)

	sugar := Sugar()
	require.NotNil(t, sugar)
}

func TestLogger_Initialized(t *testing.T) {
	err := Init(LogConfig{Level: "info"})
	require.NoError(t, err)

	logger := Logger()
	require.NotNil(t, logger)

	sugar := Sugar()
	require.NotNil(t, sugar)
}

func TestSubsystem(t *testing.T) {
	err := Init(LogConfig{Level: "info"})
	require.NoError(t, err)

	sub := Subsystem("test-subsystem")
	require.NotNil(t, sub)
}

func TestSubsystemSugar(t *testing.T) {
	err := Init(LogConfig{Level: "info"})
	require.NoError(t, err)

	sub := SubsystemSugar("test-subsystem")
	require.NotNil(t, sub)
}

func TestSync_NotInitialized(t *testing.T) {
	origRoot := rootLogger
	rootLogger = nil
	t.Cleanup(func() { rootLogger = origRoot })

	err := Sync()
	assert.NoError(t, err)
}

func TestCommonSubsystemLoggers(t *testing.T) {
	err := Init(LogConfig{Level: "info"})
	require.NoError(t, err)

	loggers := []struct {
		give string
		fn   func() interface{ Info(args ...interface{}) }
	}{
		{give: "App", fn: func() interface{ Info(args ...interface{}) } { return App() }},
		{give: "Agent", fn: func() interface{ Info(args ...interface{}) } { return Agent() }},
		{give: "Gateway", fn: func() interface{ Info(args ...interface{}) } { return Gateway() }},
		{give: "Channel", fn: func() interface{ Info(args ...interface{}) } { return Channel() }},
		{give: "Tool", fn: func() interface{ Info(args ...interface{}) } { return Tool() }},
		{give: "Session", fn: func() interface{ Info(args ...interface{}) } { return Session() }},
		{give: "Config", fn: func() interface{ Info(args ...interface{}) } { return Config() }},
	}

	for _, tt := range loggers {
		t.Run(tt.give, func(t *testing.T) {
			logger := tt.fn()
			assert.NotNil(t, logger)
		})
	}
}
