package main

import (
	"fmt"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/order/config"
	"gitlab.com/ecommercehub1/order/internal/bootstrap"
	"gitlab.com/ecommercehub1/order/internal/common/log"
	"go.uber.org/fx"
)

func init() {
	config.LoadConfig()
	logger := log.GetLogger()
	logger.Info("Order Service Config is loaded")
	errors.Init("ORDER", nil)
	logger.Info("Order Service Error is initialized")
}

func main() {
	err := config.LoadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(config.AppConfig)

	app := fx.New(
		bootstrap.DatabaseModule(),
		bootstrap.CacheModule(),
		bootstrap.RepositoryModule(),
		bootstrap.ServiceModule(),
		bootstrap.PresentModule(),
		bootstrap.TelemetryModule(),
	)


	app.Run()
}
