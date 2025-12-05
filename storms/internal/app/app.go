package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/app/configs"
	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service"
)

type App struct {
	cfg *configs.AppConfig
	svc *service.Service
}

func NewApp(appConfig *configs.AppConfig) (*App, error) {
	return &App{
		cfg: appConfig,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	ctx = a.setupSignalHandlers(ctx)

	// Start service
	a.svc = service.NewService(fmt.Sprintf("%s:%d", a.cfg.LocalIP, a.cfg.GrpcPort))
	if err := a.svc.Start(); err != nil {
		return fmt.Errorf("failed to start app service: %w", err)
	}

	// Start periodic syncing
	go func() {
		syncInterval := time.Duration(a.cfg.SyncIntervalHrs) * time.Hour
		ticker := time.NewTicker(syncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Info().Msg("Performing periodic sync")
				_, err := a.svc.SyncAllResources(ctx, &storms.SyncAllResourcesRequest{})
				if err != nil {
					log.Error().Err(err).Msg("failed to sync all resources")
				}
			}
		}
	}()

	<-ctx.Done() // Blocking call so application continues to serve.

	return nil
}

func (a *App) setupSignalHandlers(pctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(pctx)

	// Listen for interrupt signals.
	interrupt := make(chan os.Signal, 1)
	// Ctrl-C
	signal.Notify(interrupt, os.Interrupt)

	// this is what docker sends when shutting down a container
	signal.Notify(interrupt, syscall.SIGTERM)

	// Go routine to listen for interrupt signal.
	go func() {
		select {
		case <-ctx.Done():
			log.Info().Msg("App exited with Done")

			return
		case <-interrupt:
			log.Info().Msg("Interrupt signal received - shutting down StorMS")
			cancel()
		}
	}()

	return ctx
}
