package main

import (
	"context"
	"fmt"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/command"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/logginglevel"
	"go.uber.org/zap"
	"os"
	"os/signal"
)

func main() {
	if !mainImpl() {
		os.Exit(1)
	}
}

func mainImpl() bool {
	// Set up a signal-interruptible context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Initialize logger
	cfg := zap.NewProductionConfig()
	cfg.Level = logginglevel.Level
	logger, err := cfg.Build()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		return false
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Replace zap.L() and zap.S() to avoid
	// propagating the *zap.Logger by hand
	zap.ReplaceGlobals(logger)

	// Run the command
	if err := command.NewRootCmd().ExecuteContext(ctx); err != nil {
		logger.Sugar().Error(err)

		return false
	}

	return true
}
