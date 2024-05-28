package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonas27/lsd/cmd"
)

const (
	exitCodeErr = 1
)

var lvl = new(slog.LevelVar)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err := cmd.Execute(ctx, log); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "Carpet stopped with error: %v\n", err)
		os.Exit(exitCodeErr)
	}
}
