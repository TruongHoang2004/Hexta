package ai

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	aiv1 "github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/ai/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// runCallTimeout bounds one unary RPC. StartRun only accepts work; inference runs asynchronously in the AI service.
const runCallTimeout = 5 * time.Second

// errRunFailed is the sanitized transport/protocol failure; RPC details never reach business errors.
var errRunFailed = errors.New("AI run request failed")

// RunClient calls the internal Python AIService (StartRun/GetRun/CancelRun) with the shared service token.
// The token authenticates the service, not a user; Go still authorizes every business decision.
type RunClient struct {
	conn   *grpc.ClientConn
	client aiv1.AIServiceClient
	token  string
}

// NewRunClient builds the client from ai.run_addr and ai.internal_token. An empty address yields a client that
// reports types.ErrAIUnavailable, so the router hands conversations to staff instead of failing startup.
// Transport is plaintext gRPC on the private service network (as the AI service serves it); TLS is not yet
// configured for this hop.
func NewRunClient(cfg *config.Config) (*RunClient, error) {
	if cfg == nil {
		return nil, errors.New("AI configuration required")
	}
	addr := strings.TrimSpace(cfg.AI.RunAddr)
	if addr == "" {
		return &RunClient{}, nil
	}
	if len(cfg.AI.InternalToken) < 24 || strings.ContainsAny(addr, "/@?#") {
		return nil, errors.New("invalid internal AI run configuration")
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("invalid internal AI run configuration")
	}
	return &RunClient{conn: conn, client: aiv1.NewAIServiceClient(conn), token: cfg.AI.InternalToken}, nil
}

// Close releases the connection; safe on an unconfigured client.
func (c *RunClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *RunClient) call(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, runCallTimeout)
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token), cancel
}

func (c *RunClient) StartRun(ctx context.Context, input types.AIRunRequest) (types.AIRun, error) {
	if c.client == nil {
		return types.AIRun{}, types.ErrAIUnavailable
	}
	call, cancel := c.call(ctx)
	defer cancel()
	run, err := c.client.StartRun(call, startRunRequest(input))
	if status.Code(err) == codes.AlreadyExists {
		// The AI service reports ALREADY_EXISTS only when the run ID is stored with different input (a resend of the
		// same input returns the stored run). That run does not answer this request, so it is never read back.
		return types.AIRun{}, types.ErrAIRunConflict
	}
	if err != nil {
		return types.AIRun{}, errRunFailed
	}
	return runResult(run), nil
}

func (c *RunClient) GetRun(ctx context.Context, tenant, runID string) (types.AIRun, error) {
	if c.client == nil {
		return types.AIRun{}, types.ErrAIUnavailable
	}
	call, cancel := c.call(ctx)
	defer cancel()
	run, err := c.client.GetRun(call, &aiv1.RunRequest{TenantId: tenant, RunId: runID})
	if status.Code(err) == codes.NotFound {
		return types.AIRun{}, types.ErrAIRunNotFound
	}
	if err != nil {
		return types.AIRun{}, errRunFailed
	}
	return runResult(run), nil
}

func (c *RunClient) CancelRun(ctx context.Context, tenant, runID string) error {
	if c.client == nil {
		return types.ErrAIUnavailable
	}
	call, cancel := c.call(ctx)
	defer cancel()
	if _, err := c.client.CancelRun(call, &aiv1.RunRequest{TenantId: tenant, RunId: runID}); err != nil {
		return errRunFailed
	}
	return nil
}

func startRunRequest(input types.AIRunRequest) *aiv1.StartRunRequest {
	request := &aiv1.StartRunRequest{TenantId: input.TenantID, RunId: input.RunID, TurnId: input.TurnID, ContextRevision: input.ContextRevision,
		ControlGeneration: input.ControlGeneration, Message: input.Message, AllowGeneration: input.AllowGeneration}
	for _, script := range input.Scripts {
		request.Scripts = append(request.Scripts, &aiv1.Script{Id: script.ID, Trigger: script.Trigger, Text: script.Text, Active: true})
	}
	for _, item := range input.History {
		request.History = append(request.History, &aiv1.ConversationMessage{Role: item.Role, Text: item.Text})
	}
	return request
}

func runResult(run *aiv1.Run) types.AIRun {
	result := types.AIRun{RunID: run.GetRunId(), Route: run.GetRoute(), Text: run.GetText(), ScriptID: run.GetScriptId(), Reason: run.GetReason(),
		ErrorCode: run.GetErrorCode(), ContextRevision: run.GetContextRevision(), ControlGeneration: run.GetControlGeneration()}
	switch run.GetStatus() {
	case aiv1.RunStatus_RUN_STATUS_QUEUED:
		result.Status = types.AIRunQueued
	case aiv1.RunStatus_RUN_STATUS_RUNNING:
		result.Status = types.AIRunRunning
	case aiv1.RunStatus_RUN_STATUS_SUCCEEDED:
		result.Status = types.AIRunSucceeded
	case aiv1.RunStatus_RUN_STATUS_CANCELLED:
		result.Status = types.AIRunCancelled
	default:
		result.Status = types.AIRunFailed
	}
	return result
}
