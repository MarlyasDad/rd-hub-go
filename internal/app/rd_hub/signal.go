package app

import (
	"context"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func RunSignalHandler(ctx context.Context, wg *sync.WaitGroup, sugar *zap.SugaredLogger) context.Context {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	sigCtx, cancel := context.WithCancel(ctx)

	wg.Add(1)
	go func() {
		defer sugar.Info("RD-Hub: [signal] terminate")
		defer signal.Stop(sigterm)
		defer wg.Done()
		defer cancel()

		for {
			select {
			case sig, ok := <-sigterm:
				if !ok {
					sugar.Infof("RD-Hub: [signal] signal chan closed: %s\n", sig.String())
					return
				}

				sugar.Infof("RD-Hub: [signal] signal recv: %s", sig.String())
				return
			case _, ok := <-sigCtx.Done():
				if !ok {
					sugar.Info("RD-Hub: [signal] context closed")
					return
				}

				sugar.Infof("RD-Hub: [signal] ctx done: %n", sigCtx.Err().Error())
				return
			}
		}
	}()

	return sigCtx
}
