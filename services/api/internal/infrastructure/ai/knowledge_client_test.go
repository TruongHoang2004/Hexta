package ai

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	aiv1 "github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/ai/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// knowledgeServer extends the fake AIService with grounded-answer run fields and knowledge ingest.
type knowledgeServer struct {
	fakeAIServer
	upserts []*aiv1.UpsertKnowledgeRequest
	deletes []*aiv1.DeleteKnowledgeRequest
	code    codes.Code
}

func (s *knowledgeServer) GetRun(ctx context.Context, req *aiv1.RunRequest) (*aiv1.Run, error) {
	if err := s.auth(ctx); err != nil {
		return nil, err
	}
	return &aiv1.Run{TenantId: req.TenantId, RunId: req.RunId, Status: aiv1.RunStatus_RUN_STATUS_SUCCEEDED, Route: "generate", Text: "Dạ còn ạ",
		CandidateKind: "generated", CitedSourceIds: []string{"product:p1", "document:d1"}, NeedsHuman: false, HandoffReason: "",
		RequiresReview: true, Router: "jev", RouterModel: "jev-1.14.0", RouteProbability: 0.83, RouterFallbackReason: ""}, nil
}

func (s *knowledgeServer) UpsertKnowledge(ctx context.Context, req *aiv1.UpsertKnowledgeRequest) (*aiv1.UpsertKnowledgeResponse, error) {
	if err := s.auth(ctx); err != nil {
		return nil, err
	}
	s.upserts = append(s.upserts, req)
	if s.code != codes.OK {
		return nil, status.Error(s.code, "provider detail that must not leak")
	}
	return &aiv1.UpsertKnowledgeResponse{Results: []*aiv1.KnowledgeResult{{SourceType: "product", SourceId: "p1", Outcome: "created", ChunkCount: 1}}}, nil
}

func (s *knowledgeServer) DeleteKnowledge(ctx context.Context, req *aiv1.DeleteKnowledgeRequest) (*aiv1.DeleteKnowledgeResponse, error) {
	if err := s.auth(ctx); err != nil {
		return nil, err
	}
	s.deletes = append(s.deletes, req)
	if s.code != codes.OK {
		return nil, status.Error(s.code, "detail")
	}
	return &aiv1.DeleteKnowledgeResponse{Outcome: "deleted"}, nil
}

func TestRunClientMapsGroundedAnswerFieldsAndKnowledgeOptIn(t *testing.T) {
	ctx := context.Background()
	server := &knowledgeServer{}
	client := bufClient(t, server, testToken)
	if _, err := client.StartRun(ctx, types.AIRunRequest{TenantID: "t1", RunID: "r1", Message: "alo", AllowKnowledgeAnswer: true}); err != nil {
		t.Fatal(err)
	}
	if !server.started[0].AllowKnowledgeAnswer {
		t.Fatal("allow_knowledge_answer must be sent")
	}
	run, err := client.GetRun(ctx, "t1", "r1")
	if err != nil {
		t.Fatal(err)
	}
	if run.CandidateKind != "generated" || !run.RequiresReview || !reflect.DeepEqual(run.CitedSourceIDs, []string{"product:p1", "document:d1"}) ||
		run.Router != "jev" || run.RouterModel != "jev-1.14.0" || run.RouteProbability != 0.83 || run.NeedsHuman {
		t.Fatalf("grounded-answer fields not mapped: %+v", run)
	}
}

func TestKnowledgeClientSendsDocumentsAndClassifiesErrors(t *testing.T) {
	ctx := context.Background()
	server := &knowledgeServer{}
	client := bufClient(t, server, testToken)
	outcome, err := client.UpsertKnowledge(ctx, "t1", types.KnowledgeUpsert{SourceType: "product", SourceID: "p1", Revision: 7, Title: "Áo", Body: "Danh mục: Áo", Attributes: map[string]string{"sku": "A1"}})
	if err != nil || outcome != "created" {
		t.Fatal(outcome, err)
	}
	doc := server.upserts[0].Documents[0]
	if server.upserts[0].TenantId != "t1" || doc.SourceRevision != 7 || doc.Title != "Áo" || doc.Attributes["sku"] != "A1" {
		t.Fatalf("document not mapped: %+v", server.upserts[0])
	}
	if outcome, err := client.DeleteKnowledge(ctx, "t1", "script", "s1", 9); err != nil || outcome != "deleted" || server.deletes[0].SourceRevision != 9 {
		t.Fatal(outcome, err)
	}
	for code, permanent := range map[codes.Code]string{codes.InvalidArgument: "invalid_argument", codes.FailedPrecondition: "failed_precondition"} {
		server.code = code
		_, err := client.UpsertKnowledge(ctx, "t1", types.KnowledgeUpsert{SourceType: "product", SourceID: "p1"})
		var rejected *types.KnowledgeRejectedError
		if !errors.As(err, &rejected) || rejected.Code != permanent {
			t.Fatalf("%v must be a permanent rejection, got %v", code, err)
		}
	}
	for _, code := range []codes.Code{codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded, codes.Internal} {
		server.code = code
		if _, err := client.DeleteKnowledge(ctx, "t1", "product", "p1", 1); !errors.Is(err, types.ErrKnowledgeRetryable) {
			t.Fatalf("%v must be retryable, got %v", code, err)
		}
	}
	unconfigured, _ := NewRunClient(&config.Config{})
	if _, err := unconfigured.UpsertKnowledge(ctx, "t1", types.KnowledgeUpsert{}); !errors.Is(err, types.ErrAIUnavailable) {
		t.Fatal("an unconfigured client reports AI unavailable", err)
	}
}
