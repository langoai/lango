package logging

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// rootLogger is the global logger instance
	rootLogger *zap.Logger
	// sugarLogger is the sugared version for convenience
	sugarLogger *zap.SugaredLogger
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string
	Format     string
	OutputPath string
	Writer     io.Writer // explicit writer (overrides OutputPath when set)
}

// Init initializes the logging system with the given configuration
func Init(cfg LogConfig) error {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var writeSyncer zapcore.WriteSyncer
	switch {
	case cfg.Writer != nil:
		writeSyncer = zapcore.AddSync(cfg.Writer)
	case cfg.OutputPath != "":
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writeSyncer = zapcore.AddSync(file)
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	rootLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugarLogger = rootLogger.Sugar()

	return nil
}

// parseLevel converts string level to zapcore.Level
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

// Logger returns the root logger
func Logger() *zap.Logger {
	if rootLogger == nil {
		// Return a no-op logger if not initialized
		return zap.NewNop()
	}
	return rootLogger
}

// Sugar returns the sugared logger
func Sugar() *zap.SugaredLogger {
	if sugarLogger == nil {
		return zap.NewNop().Sugar()
	}
	return sugarLogger
}

// Subsystem creates a named subsystem logger
func Subsystem(name string) *zap.Logger {
	return Logger().Named(name)
}

// SubsystemSugar creates a named sugared subsystem logger
func SubsystemSugar(name string) *zap.SugaredLogger {
	return Subsystem(name).Sugar()
}

// Sync flushes any buffered log entries
func Sync() error {
	if rootLogger != nil {
		return rootLogger.Sync()
	}
	return nil
}

// Common subsystem loggers
var (
	App     = func() *zap.SugaredLogger { return SubsystemSugar("app") }
	Agent   = func() *zap.SugaredLogger { return SubsystemSugar("agent") }
	Gateway = func() *zap.SugaredLogger { return SubsystemSugar("gateway") }
	Channel = func() *zap.SugaredLogger { return SubsystemSugar("channel") }
	Tool    = func() *zap.SugaredLogger { return SubsystemSugar("tool") }
	Session = func() *zap.SugaredLogger { return SubsystemSugar("session") }
	Config  = func() *zap.SugaredLogger { return SubsystemSugar("config") }
)
