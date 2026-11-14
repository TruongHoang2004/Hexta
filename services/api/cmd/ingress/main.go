package main

import (
	"context"
	"fmt"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/bootstrap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/constant"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	apperrors.Init(constant.ServiceName, constant.TraceIdName)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if bootstrap.RunIngressProcess(ctx) != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "ingress consumer stopped; check configuration and infrastructure")
		os.Exit(1)
	}
}
