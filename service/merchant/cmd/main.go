package main

import (
	"fmt"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/merchant/config"
	"gitlab.com/ecommercehub1/merchant/internal/bootstrap"
	"go.uber.org/fx"
)

func init() {
	errors.Init("MERCHANT", nil)
	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
}

func main() {
	app := fx.New(
		bootstrap.DatabaseModule(),
		bootstrap.CacheModule(),
		bootstrap.RepositoryModule(),
		bootstrap.ServiceModule(),
		bootstrap.PresentModule(),
	)

	app.Run()
}
