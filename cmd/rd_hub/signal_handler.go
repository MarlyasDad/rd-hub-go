package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func RunSignalHandler(ctx context.Context, wg *sync.WaitGroup) context.Context {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	sigCtx, cancel := context.WithCancel(ctx)

	wg.Add(1)
	go func() {
		defer slog.Info("RD-Hub: [signal] terminate")
		defer signal.Stop(sigterm)
		defer wg.Done()
		defer cancel()

		for {
			select {
			case sig, ok := <-sigterm:
				if !ok {
					slog.Info(fmt.Sprintf("RD-Hub: [signal] signal chan closed: %s\n", sig.String()))
					return
				}

				slog.Info(fmt.Sprintf("RD-Hub: [signal] signal recv: %s", sig.String()))
				return
			case _, ok := <-sigCtx.Done():
				if !ok {
					slog.Info("RD-Hub: [signal] context closed")
					return
				}

				slog.Info(fmt.Sprintf("RD-Hub: [signal] ctx done: %s", sigCtx.Err().Error()))
				return
			}
		}
	}()

	return sigCtx
}
