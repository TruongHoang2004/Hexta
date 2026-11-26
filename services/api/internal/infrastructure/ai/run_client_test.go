package ai

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	aiv1 "github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/ai/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const testToken = "test-internal-token-0123456789"

// fakeAIServer is an in-process AIService: it checks the bearer token and records requests.
type fakeAIServer struct {
	aiv1.UnimplementedAIServiceServer
	started  []*aiv1.StartRunRequest
	conflict bool
	missing  bool
	polled   int
	status   aiv1.RunStatus
}

func (s *fakeAIServer) auth(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)
	if values := md.Get("authorization"); len(values) != 1 || values[0] != "Bearer "+testToken {
		return status.Error(codes.Unauthenticated, "bad token")
	}
	return nil
}

func (s *fakeAIServer) StartRun(ctx context.Context, req *aiv1.StartRunRequest) (*aiv1.Run, error) {
	if err := s.auth(ctx); err != nil {
		return nil, err
	}
	s.started = append(s.started, req)
	if s.conflict {
		return nil, status.Error(codes.AlreadyExists, "Run ID has different input")
	}
	return &aiv1.Run{TenantId: req.TenantId, RunId: req.RunId, Status: aiv1.RunStatus_RUN_STATUS_QUEUED, ContextRevision: req.ContextRevision, ControlGeneration: req.ControlGeneration}, nil
}

func (s *fakeAIServer) GetRun(ctx context.Context, req *aiv1.RunRequest) (*aiv1.Run, error) {
	if err := s.auth(ctx); err != nil {
		return nil, err
	}
	s.polled++
	if s.missing {
		return nil, status.Error(codes.NotFound, "Run not found")
	}
	return &aiv1.Run{TenantId: req.TenantId, RunId: req.RunId, Status: s.status, Route: "clarify", Text: "Bạn hỏi mẫu nào?", ContextRevision: 2, ControlGeneration: 1}, nil
}

func (s *fakeAIServer) CancelRun(ctx context.Context, req *aiv1.RunRequest) (*aiv1.Run, error) {
	if err := s.auth(ctx); err != nil {
		return nil, err
	}
	return &aiv1.Run{RunId: req.RunId, Status: aiv1.RunStatus_RUN_STATUS_CANCELLED}, nil
}

func bufClient(t *testing.T, server aiv1.AIServiceServer, token string) *RunClient {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	aiv1.RegisterAIServiceServer(srv, server)
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(srv.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &RunClient{conn: conn, client: aiv1.NewAIServiceClient(conn), token: token}
}

func TestRunClientMapsContractAndAuthenticates(t *testing.T) {
	ctx := context.Background()
	server := &fakeAIServer{status: aiv1.RunStatus_RUN_STATUS_SUCCEEDED}
	client := bufClient(t, server, testToken)
	run, err := client.StartRun(ctx, types.AIRunRequest{TenantID: "t1", RunID: "r1", TurnID: "turn", ContextRevision: 2, ControlGeneration: 1, Message: "alo",
		Scripts: []types.AIScript{{ID: "v1", Trigger: "giá", Text: "200k"}}, History: []types.AIHistoryItem{{Role: "user", Text: "hi"}}, AllowGeneration: true})
	if err != nil || run.Status != types.AIRunQueued || run.RunID != "r1" {
		t.Fatal(run, err)
	}
	req := server.started[0]
	if !req.AllowGeneration || len(req.Scripts) != 1 || !req.Scripts[0].Active || req.Scripts[0].Id != "v1" || len(req.History) != 1 || req.History[0].Role != "user" {
		t.Fatalf("request not mapped: %+v", req)
	}
	run, err = client.GetRun(ctx, "t1", "r1")
	if err != nil || run.Status != types.AIRunSucceeded || run.Route != "clarify" || run.ContextRevision != 2 {
		t.Fatal(run, err)
	}
	// ALREADY_EXISTS means the run ID holds different input: a conflict, never a read of that other run.
	server.conflict = true
	polled := server.polled
	if run, err = client.StartRun(ctx, types.AIRunRequest{TenantID: "t1", RunID: "r1", Message: "alo"}); !errors.Is(err, types.ErrAIRunConflict) || run.Status != "" {
		t.Fatal("ALREADY_EXISTS must map to a run conflict", run, err)
	}
	if server.polled != polled {
		t.Fatal("a conflicting run must not be read back")
	}
	// NOT_FOUND is distinguishable so the router can start the run again.
	server.missing = true
	if _, err = client.GetRun(ctx, "t1", "r1"); !errors.Is(err, types.ErrAIRunNotFound) {
		t.Fatal("NOT_FOUND must map to run not found", err)
	}
	server.missing = false
	if err := client.CancelRun(ctx, "t1", "r1"); err != nil {
		t.Fatal(err)
	}
	server.status = aiv1.RunStatus_RUN_STATUS_UNSPECIFIED
	if run, _ = client.GetRun(ctx, "t1", "r1"); run.Status != types.AIRunFailed {
		t.Fatal("unknown run status must fail closed")
	}
	// Transport errors are sanitized.
	bad := bufClient(t, &fakeAIServer{}, "wrong-token-0123456789-abcdef")
	if _, err := bad.GetRun(ctx, "t1", "r1"); !errors.Is(err, errRunFailed) {
		t.Fatal("RPC errors must be sanitized", err)
	}
}

func TestRunClientUnconfiguredAndInvalidConfig(t *testing.T) {
	cfg := &config.Config{}
	client, err := NewRunClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.StartRun(context.Background(), types.AIRunRequest{}); !errors.Is(err, types.ErrAIUnavailable) {
		t.Fatal("an unconfigured client must report AI unavailable", err)
	}
	if client.Close() != nil {
		t.Fatal("closing an unconfigured client must succeed")
	}
	cfg.AI.RunAddr = "ai:50051"
	cfg.AI.InternalToken = "short"
	if _, err := NewRunClient(cfg); err == nil {
		t.Fatal("a short service token must fail startup")
	}
	cfg.AI.InternalToken = testToken
	configured, err := NewRunClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_ = configured.Close()
}
