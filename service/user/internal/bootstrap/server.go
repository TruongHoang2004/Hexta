package bootstrap

import (
	"context"
	"fmt"
	stdlog "log"
	"net"

	"gitlab.com/ecommercehub1/lib/pkg/telemetry"
	"gitlab.com/ecommercehub1/user/config"
	"gitlab.com/ecommercehub1/user/internal/common/log"
	present "gitlab.com/ecommercehub1/user/internal/present/handler"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcRegistrar interface allows handlers to self-register
type GrpcRegistrar interface {
	Register(*grpc.Server)
}

// AsGrpcRegistrar is a helper to annotate handlers for the fx group
func AsGrpcRegistrar(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(GrpcRegistrar)),
		fx.ResultTags(`group:"grpc_handlers"`),
	)
}

func PresentModule() fx.Option {
	return fx.Options(
		fx.Provide(
			AsGrpcRegistrar(present.NewGrpcHandler),
			AsGrpcRegistrar(present.NewAddressGrpcHandler),
			AsGrpcRegistrar(present.NewMerchantGrpcHandler),
		),
		fx.Invoke(StartGrpcServer),
	)
}

type GrpcServerParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Handlers  []GrpcRegistrar `group:"grpc_handlers"`
}

func StartGrpcServer(params GrpcServerParams) {
	var grpcServer *grpc.Server
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.AppConfig.Server.Port))
			if err != nil {
				return fmt.Errorf("failed to listen: %w", err)
			}

			grpcServer = grpc.NewServer(
				grpc.StatsHandler(otelgrpc.NewServerHandler()),
				grpc.ChainUnaryInterceptor(
					telemetry.GrpcInterceptor(log.GetLogger().GetZap()),
				),
			)
			reflection.Register(grpcServer)

			// Register all handlers in the group
			for _, handler := range params.Handlers {
				handler.Register(grpcServer)
			}

			go func() {
				stdlog.Printf("🚀 gRPC server listening on %s", config.AppConfig.Server.Port)
				if err := grpcServer.Serve(lis); err != nil {
					stdlog.Fatalf("failed to serve: %v", err)
				}
			}()

			go func() {
				metricsPort := "9091" // Or get from config
				stdlog.Printf("📊 Metrics server listening on %s", metricsPort)
				if err := telemetry.ExposeMetrics(metricsPort); err != nil {
					stdlog.Printf("failed to expose metrics: %v", err)
				}
			}()

			return nil

		},
		OnStop: func(ctx context.Context) error {
			stdlog.Println("🛑 Stopping gRPC server...")
			if grpcServer != nil {
				grpcServer.GracefulStop()
			}
			return nil
		},
	})
}
