package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/ngoldack/dicetrace/apps/bgg-proxy/internal"
	"github.com/ngoldack/dicetrace/package/core/logger"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := Run(context.Background()); err != nil {
		panic(err)
	}
}

const instances = 3

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger.SetupLogger()

	slog.Info("starting bgg-proxy...")

	errg, ctx := errgroup.WithContext(ctx)

	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer nc.Close()

	srvs := make(map[string]micro.Service)
	mu := sync.RWMutex{}

	// Create 3 instances of the service
	for range instances {
		// Register handlers
		errg.Go(func() error {
			srv, err := micro.AddService(nc, micro.Config{
				// Add service configuration here
				Name:    "bgg-proxy",
				Version: "1.0.0",
			})
			if err != nil {
				return fmt.Errorf("failed to create micro service: %w", err)
			}

			err = srv.AddEndpoint("bgg-game-by-id", internal.HandlerGetGameByID())
			if err != nil {
				return fmt.Errorf("failed to add GetGameByID endpoint: %w", err)
			}

			mu.Lock()
			srvs[srv.Info().ID] = srv
			mu.Unlock()

			slog.Info("starting micro service", slog.Any("info", srv.Info()))

			return nil
		})
	}

	mu.RLock()
	for _, srv := range srvs {
		errg.Go(func() error {
			<-ctx.Done()

			slog.Info("shutting down service", slog.Any("info", srv.Info()))

			if err := srv.Stop(); err != nil {
				slog.Error("failed to stop micro service", "error", err)
			}

			mu.Lock()
			delete(srvs, srv.Info().ID)
			mu.Unlock()

			return nil
		})
	}
	mu.RUnlock()

	// Blocking go-routine to wait for context cancellation
	errg.Go(func() error {
		<-ctx.Done()
		slog.Info("context cancelled, shutting down bgg-proxy...")
		return nil
	})

	err = errg.Wait()
	if err != nil {
		return fmt.Errorf("bgg-proxy exited with error: %w", err)
	}

	slog.Info("bgg-proxy exited gracefully")

	return nil
}
