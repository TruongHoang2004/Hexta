package main

import (
	"context"
	"fmt"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/bootstrap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if bootstrap.Run(ctx) != nil {
		fmt.Fprintln(os.Stderr, "connector stopped: check configuration and infrastructure")
		os.Exit(1)
	}
}
