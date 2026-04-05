package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net"

	"gitlab.com/ecommercehub1/user/config"
	present "gitlab.com/ecommercehub1/user/internal/present/handler"
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

			grpcServer = grpc.NewServer()
			reflection.Register(grpcServer)

			// Register all handlers in the group
			for _, handler := range params.Handlers {
				handler.Register(grpcServer)
			}

			go func() {
				log.Printf("🚀 gRPC server listening on %s", config.AppConfig.Server.Port)
				if err := grpcServer.Serve(lis); err != nil {
					log.Fatalf("failed to serve: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("🛑 Stopping gRPC server...")
			if grpcServer != nil {
				grpcServer.GracefulStop()
			}
			return nil
		},
	})
}
