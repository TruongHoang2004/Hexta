// Command messaging-command serves the conversation write side (claim, assign, handoff, resolve, replies, simulator
// input, bot proposal send/discard/regenerate) and runs the 24-hour turn expiry job, the bot reply router and the AI
// knowledge sync job.
// The public API forwards those routes here and keeps serving reads; both share the API database.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/bootstrap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/constant"
	"github.com/shopspring/decimal"
	"go.uber.org/fx"
)

const gracefulTimeout = 15 * time.Second

func init() {
	errors.Init(constant.ServiceName, constant.TraceIdName)
	decimal.MarshalJSONWithoutQuotes = true
}

func main() {
	if err := config.LoadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Configuration validation failed:", err)
		os.Exit(1)
	}
	if port := config.AppConfig.Messaging.CommandPort; port != "" {
		config.AppConfig.Server.Port = port
	}
	logger := log.GetLogger().GetZap()

	app := fx.New(
		fx.Provide(log.GetLogger().GetZap),
		bootstrap.BuildDatabase(),
		bootstrap.BuildCache(),
		bootstrap.BuildRepository(),
		bootstrap.BuildService(),
		bootstrap.BuildController(),
		bootstrap.BuildValidator(),
		bootstrap.BuildMiddleware(),
		bootstrap.ServerModule,
		bootstrap.MessagingCommandRouterModule,
		// Conversation writes live here, so the 24-hour turn expiry job does too.
		bootstrap.BuildTurnExpiry(),
		// The bot reply router turns customer messages on bot-handled turns into proposals or bot sends.
		bootstrap.BuildReplyRouter(),
		// AI knowledge sync (products, active scripts, knowledge documents) shares the router's AI client; it must
		// follow BuildReplyRouter so it stops before the client is closed.
		bootstrap.BuildKnowledgeSync(),
	)

	startCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		logger.Fatalf(err.Error())
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		logger.Fatalf(err.Error())
	}
}
