// Package logging provides centralized logging functionality.
package logging

import (
	"fmt"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehhttp"
	"github.com/lefinal/meh/mehlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
	"sync"
)

func init() {
	mehlog.OmitErrorMessageField(true)
}

// Encoding of log entries to use.
type Encoding string

const (
	// EncodingConsole encodes log entries for human-readability in console output
	// with colored log levels.
	EncodingConsole = "console"
	// EncodingJSON encodes log entries as JSON for better machine-readability.
	EncodingJSON = "json"
)

// ParseEncoding parses the given string representation as Encoding.
func ParseEncoding(s string) (Encoding, error) {
	switch strings.ToLower(s) {
	case "console":
		return EncodingConsole, nil
	case "json":
		return EncodingJSON, nil
	default:
		return "", fmt.Errorf("unknown log encoding: %s", s)
	}
}

// NewLogger creates a new zap.Logger for the given log level and encoding. Don't
// forget to call Sync() on the returned logged before exiting!
func NewLogger(level zapcore.Level, encoding Encoding) (*zap.Logger, error) {
	switch encoding {
	case EncodingConsole:
		return newConsoleLogger(level)
	case EncodingJSON:
		return newJSONLogger(level)
	default:
		return nil, fmt.Errorf("unsupported log encoding: %s", encoding)
	}
}

// newConsoleLogger creates a new zap.Logger.
func newConsoleLogger(level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	//nolint:exhaustruct
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	config.OutputPaths = []string{"stdout"}
	config.Level = zap.NewAtomicLevelAt(level)
	config.DisableCaller = true
	config.DisableStacktrace = true
	logger, err := config.Build()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "new zap production logger", meh.Details{"config": config})
	}
	return logger, nil
}

// newJSONLogger creates a new zap.Logger that formats log entries via JSON.
func newJSONLogger(level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	//nolint:exhaustruct
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	config.OutputPaths = []string{"stdout"}
	config.Level = zap.NewAtomicLevelAt(level)
	config.DisableCaller = true
	config.DisableStacktrace = true
	logger, err := config.Build()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "new zap production logger", meh.Details{"config": config})
	}
	return logger, nil
}

var (
	logger      *zap.Logger
	loggerMutex sync.RWMutex
)

var defaultLevelTranslator map[meh.Code]zapcore.Level
var defaultLevelTranslatorMutex sync.RWMutex

func init() {
	defaultLevelTranslator = make(map[meh.Code]zapcore.Level)
	AddToDefaultLevelTranslator(meh.ErrNotFound, zap.DebugLevel)
	AddToDefaultLevelTranslator(meh.ErrUnauthorized, zap.DebugLevel)
	AddToDefaultLevelTranslator(meh.ErrForbidden, zap.DebugLevel)
	AddToDefaultLevelTranslator(meh.ErrBadInput, zap.DebugLevel)
	AddToDefaultLevelTranslator(mehhttp.ErrCommunication, zap.DebugLevel)
	mehlog.SetDefaultLevelTranslator(func(code meh.Code) zapcore.Level {
		defaultLevelTranslatorMutex.RLock()
		defer defaultLevelTranslatorMutex.RUnlock()
		if level, ok := defaultLevelTranslator[code]; ok {
			return level
		}
		return zap.ErrorLevel
	})
}

// AddToDefaultLevelTranslator adds the given case to the translation map.
func AddToDefaultLevelTranslator(code meh.Code, level zapcore.Level) {
	defaultLevelTranslatorMutex.Lock()
	defaultLevelTranslator[code] = level
	defaultLevelTranslatorMutex.Unlock()
}

// DebugLogger returns the debug logger from SetLogger. If none is set, a new one
// will be created.
func DebugLogger() *zap.Logger {
	return RootLogger().Named("debug")
}

// RootLogger returns the logger set via SetLogger. If none is set, a new one
// will be created.
func RootLogger() *zap.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	if logger == nil {
		logger, _ = newConsoleLogger(zap.InfoLevel)
	}
	return logger
}

// SetLogger sets the logger that is used for reporting errors in main as well as
// with DebugLogger.
func SetLogger(newLogger *zap.Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	logger = newLogger
}
