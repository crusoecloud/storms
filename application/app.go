package application

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"gitlab.com/crusoeenergy/island/storage/storms/configs"
	"gitlab.com/crusoeenergy/island/storage/storms/service"
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
