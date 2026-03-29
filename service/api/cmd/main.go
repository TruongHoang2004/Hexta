package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopspring/decimal"
	"gitlab.com/ecommercehub1/api/common/log"
	"gitlab.com/ecommercehub1/api/config"
	"gitlab.com/ecommercehub1/api/internal/bootstrap"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	defaultGracefulTimeout = 15 * time.Second
)

func init() {
	config.LoadConfig()
	logger := log.GetLogger()
	logger.Info("Config is loaded")
	errors.Init("API", nil)
	logger.Info("Error is initialized")
	decimal.MarshalJSONWithoutQuotes = true
}

func main() {

	logger := log.GetLogger().GetZap()
	logger.Debugf("App is running")

	app := fx.New(
		// fx.NopLogger, // Disable Fx's own logging
		fx.Provide(log.GetLogger().GetZap),
		bootstrap.BuildDatabase(),
		bootstrap.BuildCache(),
		bootstrap.BuildClient(),
		bootstrap.BuildRepository(),
		bootstrap.BuildService(),
		bootstrap.BuildController(),
		bootstrap.BuildValidator(),
		bootstrap.BuildMiddleware(),
		bootstrap.ServerModule,
		bootstrap.RouterModule,
	)

	startCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulTimeout)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		logger.Fatalf(err.Error())
	}

	interruptHandle(app, logger)
}

func OnStart(ctx context.Context) error {
	log.Info(ctx, "Application is starting...")
	return nil
}

func OnStop(ctx context.Context) error {
	log.Info(ctx, "Application is stopping...")
	return nil
}

func interruptHandle(app *fx.App, logger *zap.SugaredLogger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger.Debugf("Listening Signal...")
	s := <-c
	logger.Infof("Received signal: %s. Shutting down Server ...", s)

	stopCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulTimeout)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		logger.Fatalf(err.Error())
	}
}
