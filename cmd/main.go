package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alanshaw/buff/cmd/cli"
)

func main() {
	// Create a context that cancels on interrupt signals
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Execute with the signal-aware context
	cli.ExecuteContext(ctx)
}
