package main

import (
	"context"
	"fmt"
	"github.com/lefinal/image-to-ma3-scribble/app"
	"github.com/lefinal/image-to-ma3-scribble/logging"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	envLogLevel          = "LOG_LEVEL"
	envHTTPAPIListenAddr = "HTTP_API_LISTEN_ADDR"
	envPotraceFilename   = "POTRACE_FILENAME"
)

func run() error {
	// Parse config.
	var config app.Config
	logLevelStr := os.Getenv(envLogLevel)
	if logLevelStr == "" {
		logLevelStr = "info"
	}
	logLevel, err := zapcore.ParseLevel(logLevelStr)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}
	logger, err := logging.NewLogger(logLevel, logging.EncodingConsole)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	config.Logger = logger
	if v := os.Getenv(envHTTPAPIListenAddr); v == "" {
		return fmt.Errorf("environment variable %s is not set", envHTTPAPIListenAddr)
	} else {
		config.HTTPAPIListenAddr = v
	}
	if v := os.Getenv(envPotraceFilename); v == "" {
		return fmt.Errorf("environment variable %s is not set", envPotraceFilename)
	} else {
		config.PotraceFilename = v
	}

	// Run.
	appInstance := app.New(config)
	err = runUntilTerminated(appInstance.Run)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	errorLogger, err := logging.NewLogger(zapcore.ErrorLevel, logging.EncodingJSON)
	if err != nil {
		log.Fatalf("create logger: %s", err.Error())
	}
	logging.SetLogger(errorLogger)

	err = run()
	if err != nil {
		errorLogger.Fatal(err.Error(), zap.Error(err))
	}
}

// waitForTerminate until a terminate-signal is received or the given context is
// done.
func waitForTerminate(ctx context.Context) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	select {
	case <-ctx.Done():
	case <-signals:
	}
}

// run the given function until a syscall.SIGINT or syscall.SIGTERM is received.
func runUntilTerminated(runnable func(ctx context.Context) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer cancel()
		waitForTerminate(ctx)
	}()
	err := runnable(ctx)
	if err != nil {
		select {
		case <-ctx.Done():
			if strings.Contains(err.Error(), "context canceled") {
				mehlog.LogToLevel(logging.RootLogger(), zapcore.DebugLevel, meh.Wrap(err, "run", meh.Details{"application_exit": true}))
				return nil
			}
		default:
		}
		return err
	}
	return nil
}
