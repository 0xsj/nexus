package zap

import (
	"os"
	"time"

	"github.com/0xsj/nexus/pkg/observability/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// buildZapConfig converts our generic logger.Config to zap.Config.
func buildZapConfig(cfg logger.Config) zap.Config {
	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(convertLevel(cfg.Level)),
		Development:       cfg.Development,
		DisableCaller:     cfg.DisableCaller,
		DisableStacktrace: cfg.DisableStacktrace,
		Encoding:          cfg.Encoding,
		EncoderConfig:     buildEncoderConfig(cfg),
		OutputPaths:       cfg.OutputPaths,
		ErrorOutputPaths:  cfg.ErrorOutputPaths,
	}

	// Configure sampling if provided
	if cfg.Sampling != nil {
		zapCfg.Sampling = &zap.SamplingConfig{
			Initial:    cfg.Sampling.Initial,
			Thereafter: cfg.Sampling.Thereafter,
		}
	}

	// Add initial fields
	if len(cfg.InitialFields) > 0 {
		zapCfg.InitialFields = cfg.InitialFields
	}

	return zapCfg
}

func buildEncoderConfig(cfg logger.Config) zapcore.EncoderConfig {
	// Determine if we should enable colors
	colorEnabled := shouldEnableColor(cfg)

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	// Configure time encoding - use a custom compact format
	if colorEnabled {
		encoderCfg.EncodeTime = compactTimeEncoder
	} else {
		encoderCfg.EncodeTime = timeEncoder(cfg.TimeEncoder)
	}

	// Configure level encoding
	if colorEnabled {
		encoderCfg.EncodeLevel = compactColorLevelEncoder
	} else if cfg.Encoding == "console" {
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	} else {
		encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	// Configure caller encoding - more compact
	if colorEnabled {
		encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	} else {
		encoderCfg.EncodeCaller = callerEncoder(cfg.CallerEncoder)
	}

	// Configure name encoding
	if colorEnabled {
		encoderCfg.EncodeName = zapcore.FullNameEncoder
	} else {
		encoderCfg.EncodeName = zapcore.FullNameEncoder
	}

	// Console-specific settings - make it compact
	if cfg.Encoding == "console" {
		encoderCfg.ConsoleSeparator = " " // Single space instead of double
	}

	return encoderCfg
}

// compactTimeEncoder formats time in HH:MM:SS.mmm format (compact like your screenshot)
func compactTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000"))
}

// compactColorLevelEncoder encodes log levels with colors in a compact format
func compactColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var color string
	var levelStr string

	switch level {
	case zapcore.DebugLevel:
		color = "\033[36m" // Cyan
		levelStr = "DEBG"
	case zapcore.InfoLevel:
		color = "\033[32m" // Green
		levelStr = "INFO"
	case zapcore.WarnLevel:
		color = "\033[33m" // Yellow
		levelStr = "WARN"
	case zapcore.ErrorLevel:
		color = "\033[31m" // Red
		levelStr = "ERRO"
	case zapcore.FatalLevel:
		color = "\033[35m" // Magenta
		levelStr = "FATL"
	case zapcore.PanicLevel:
		color = "\033[35m" // Magenta
		levelStr = "PANC"
	default:
		color = logger.ResetColor()
		levelStr = level.CapitalString()
	}

	enc.AppendString(color + levelStr + logger.ResetColor())
}

// convertLevel converts our logger.Level to zapcore.Level.
func convertLevel(level logger.Level) zapcore.Level {
	switch level {
	case logger.DebugLevel:
		return zapcore.DebugLevel
	case logger.InfoLevel:
		return zapcore.InfoLevel
	case logger.WarnLevel:
		return zapcore.WarnLevel
	case logger.ErrorLevel:
		return zapcore.ErrorLevel
	case logger.FatalLevel:
		return zapcore.FatalLevel
	case logger.PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// timeEncoder returns the appropriate time encoder based on the configuration.
func timeEncoder(encoder string) zapcore.TimeEncoder {
	switch encoder {
	case "iso8601", "":
		return zapcore.ISO8601TimeEncoder
	case "millis":
		return zapcore.EpochMillisTimeEncoder
	case "nanos":
		return zapcore.EpochNanosTimeEncoder
	case "epoch":
		return zapcore.EpochTimeEncoder
	case "rfc3339":
		return zapcore.RFC3339TimeEncoder
	case "rfc3339nano":
		return zapcore.RFC3339NanoTimeEncoder
	default:
		return zapcore.ISO8601TimeEncoder
	}
}

// callerEncoder returns the appropriate caller encoder based on the configuration.
func callerEncoder(encoder string) zapcore.CallerEncoder {
	switch encoder {
	case "short", "":
		return zapcore.ShortCallerEncoder
	case "full":
		return zapcore.FullCallerEncoder
	default:
		return zapcore.ShortCallerEncoder
	}
}

// colorLevelEncoder encodes log levels with ANSI colors for console output.
func colorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var color string
	var levelStr string

	switch level {
	case zapcore.DebugLevel:
		color = logger.DebugLevel.Color()
		levelStr = "DEBUG"
	case zapcore.InfoLevel:
		color = logger.InfoLevel.Color()
		levelStr = "INFO "
	case zapcore.WarnLevel:
		color = logger.WarnLevel.Color()
		levelStr = "WARN "
	case zapcore.ErrorLevel:
		color = logger.ErrorLevel.Color()
		levelStr = "ERROR"
	case zapcore.FatalLevel:
		color = logger.FatalLevel.Color()
		levelStr = "FATAL"
	case zapcore.PanicLevel:
		color = logger.PanicLevel.Color()
		levelStr = "PANIC"
	default:
		color = logger.ResetColor()
		levelStr = level.CapitalString()
	}

	enc.AppendString(color + levelStr + logger.ResetColor())
}

// colorNameEncoder encodes logger names with color for console output.
func colorNameEncoder(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
	if loggerName == "" {
		return
	}
	// Cyan color for logger names
	enc.AppendString("\033[36m" + loggerName + logger.ResetColor())
}

// isTerminal checks if stdout is a terminal.
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	// Check if it's a character device (terminal)
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// shouldEnableColor determines if color output should be enabled.
func shouldEnableColor(cfg logger.Config) bool {
	if !cfg.ColorEnabled {
		return false
	}

	// Only enable colors for console encoding
	if cfg.Encoding != "console" {
		return false
	}

	// Check if stdout is a terminal
	for _, path := range cfg.OutputPaths {
		if path == "stdout" || path == "stderr" {
			return isTerminal()
		}
	}

	return false
}

// buildOptions creates zap.Option slice from logger.Config.
func buildOptions(cfg logger.Config) []zap.Option {
	var opts []zap.Option

	// Add caller information
	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
		opts = append(opts, zap.AddCallerSkip(1)) // Skip wrapper functions
	}

	// Add stacktrace
	if !cfg.DisableStacktrace {
		stacktraceLevel := convertLevel(cfg.StacktraceLevel)
		opts = append(opts, zap.AddStacktrace(stacktraceLevel))
	}

	// Development mode options
	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	return opts
}
