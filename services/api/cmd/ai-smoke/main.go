// Command ai-smoke verifies the local Go -> Python gRPC simulation boundary.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	aiv1 "github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/ai/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	if err := smoke(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func smoke() error {
	token := os.Getenv("DEMO_AI_INTERNAL_TOKEN")
	if len(token) < 24 {
		return fmt.Errorf("set DEMO_AI_INTERNAL_TOKEN (at least 24 characters)")
	}
	// Plaintext transport is restricted to this fixed loopback demo endpoint.
	conn, err := grpc.NewClient("127.0.0.1:55051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := aiv1.NewAIServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	input := &aiv1.StartRunRequest{TenantId: "demo", RunId: fmt.Sprintf("smoke-%d", time.Now().UnixNano()),
		TurnId: "turn-demo", ContextRevision: 1, ControlGeneration: 1, Message: "price",
		Scripts: []*aiv1.Script{{Id: "script-demo", Trigger: "price", Text: "Please select a product.", Active: true}}}
	if _, err := client.StartRun(ctx, input); status.Code(err) != codes.Unauthenticated {
		return fmt.Errorf("unauthenticated request was not rejected: %v", err)
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	run, err := client.StartRun(ctx, input)
	if err != nil {
		return err
	}
	duplicate, err := client.StartRun(ctx, input)
	if err != nil {
		return err
	}
	if duplicate.RunId != run.RunId {
		return fmt.Errorf("idempotency mismatch")
	}
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		run, err = client.GetRun(ctx, &aiv1.RunRequest{TenantId: "demo", RunId: input.RunId})
		if err != nil {
			return err
		}
		if run.Status == aiv1.RunStatus_RUN_STATUS_SUCCEEDED {
			break
		}
		if run.Status == aiv1.RunStatus_RUN_STATUS_FAILED || run.Status == aiv1.RunStatus_RUN_STATUS_CANCELLED {
			return fmt.Errorf("unexpected terminal status: %s", run.Status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	if !run.Simulated || run.Route != "script" || run.Text != "Please select a product." || run.ContextRevision != 1 || run.ControlGeneration != 1 {
		return fmt.Errorf("unexpected candidate")
	}
	fmt.Println("Go -> Python gRPC smoke passed (auth, idempotency, simulated candidate)")
	return nil
}
