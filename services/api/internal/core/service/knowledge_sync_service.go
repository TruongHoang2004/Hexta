package service

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/shopspring/decimal"
)

// IKnowledgeAIClient is the internal AI knowledge ingest capability (Python AIService UpsertKnowledge /
// DeleteKnowledge). Both calls are idempotent per source revision, so a retry with the same revision is safe.
type IKnowledgeAIClient interface {
	UpsertKnowledge(ctx context.Context, tenant string, document types.KnowledgeUpsert) (string, error)
	DeleteKnowledge(ctx context.Context, tenant, sourceType, sourceID string, revision int64) (string, error)
}

// knowledgeGoneRevision is the removal revision of a source whose row no longer exists: IDs are never reused, so
// the tombstone may outrank every earlier and later revision.
const knowledgeGoneRevision = math.MaxInt64

// knowledgeMaxVariants bounds the product attributes (three per variant) below the AI contract's 30 attributes.
const knowledgeMaxVariants = 10

// KnowledgeSyncService pushes changed knowledge sources (products, active scripts, knowledge documents) to the AI
// service. Business writes only enqueue a durable job in their own transaction; this job reads the source's
// current state after commit and sends it with a monotonic revision. UNAVAILABLE, RESOURCE_EXHAUSTED and other
// transient failures retry later with the same revision; INVALID_ARGUMENT and FAILED_PRECONDITION are permanent
// and recorded on the job until the source changes again. Go remains the authority: Python only stores what Go
// sends and every grounded answer still waits for staff review.
type KnowledgeSyncService struct {
	store repository.KnowledgeSyncStore
	ai    IKnowledgeAIClient
}

func NewKnowledgeSyncService(store repository.KnowledgeSyncStore, ai IKnowledgeAIClient) *KnowledgeSyncService {
	return &KnowledgeSyncService{store: store, ai: ai}
}

// SyncOne processes the next due job. It reports whether a job was claimed, so callers can drain without waiting.
func (s *KnowledgeSyncService) SyncOne(ctx context.Context) (bool, *apperrors.Error) {
	job, err := s.store.ClaimKnowledgeJob(ctx)
	if err != nil || job == nil {
		return false, err
	}
	source, err := s.source(ctx, *job)
	if err != nil {
		// A source that cannot be read is retried; the lease expires or the retry is scheduled.
		if _, retryErr := s.store.RetryKnowledgeJob(ctx, *job); retryErr != nil {
			return true, retryErr
		}
		return true, err
	}
	var callErr error
	if source.Upsert != nil {
		_, callErr = s.ai.UpsertKnowledge(ctx, job.TenantID, *source.Upsert)
	} else {
		_, callErr = s.ai.DeleteKnowledge(ctx, job.TenantID, job.SourceType, job.SourceID, source.Revision)
	}
	var rejected *types.KnowledgeRejectedError
	switch {
	case callErr == nil:
		_, err = s.store.FinishKnowledgeJob(ctx, *job, source.Revision)
	case errors.As(callErr, &rejected):
		_, err = s.store.FailKnowledgeJob(ctx, *job, rejected.Code)
	default:
		// Retryable, transport failure or AI service not configured: same revision on the next attempt.
		_, err = s.store.RetryKnowledgeJob(ctx, *job)
		if err == nil {
			err = apperrors.ErrSystemError(ctx, "knowledge sync failed; retry scheduled")
		}
	}
	return true, err
}

// source reads the job's source and builds the AI document (or the removal) from its current state.
func (s *KnowledgeSyncService) source(ctx context.Context, job types.KnowledgeJob) (types.KnowledgeSource, *apperrors.Error) {
	switch job.SourceType {
	case types.KnowledgeSourceProduct:
		product, err := s.store.KnowledgeProduct(ctx, job.TenantID, job.SourceID)
		if err != nil {
			return types.KnowledgeSource{}, err
		}
		return productKnowledge(product), nil
	case types.KnowledgeSourceScript:
		script, err := s.store.KnowledgeScript(ctx, job.TenantID, job.SourceID)
		if err != nil {
			return types.KnowledgeSource{}, err
		}
		return scriptKnowledge(script), nil
	case types.KnowledgeSourceDocument:
		document, err := s.store.KnowledgeDocument(ctx, job.TenantID, job.SourceID)
		if err != nil {
			return types.KnowledgeSource{}, err
		}
		return documentKnowledge(document), nil
	}
	return types.KnowledgeSource{}, apperrors.ErrSystemError(ctx, "unknown knowledge source type")
}

// productKnowledge embeds name, category, description and variants; SKU, price (only when provided) and available
// stock are attributes, which the AI service shows to the model without re-embedding when only they change.
func productKnowledge(product *types.KnowledgeProduct) types.KnowledgeSource {
	if product == nil {
		return types.KnowledgeSource{Revision: knowledgeGoneRevision}
	}
	if product.Removed {
		return types.KnowledgeSource{Revision: product.Revision}
	}
	lines := []string{}
	if product.Category != "" {
		lines = append(lines, "Danh mục: "+product.Category)
	}
	if product.Description != "" {
		lines = append(lines, "Mô tả: "+product.Description)
	}
	attributes := map[string]string{}
	for i, variant := range product.Variants {
		if i == knowledgeMaxVariants {
			break
		}
		suffix := ""
		if i > 0 {
			suffix = "_" + strconv.Itoa(i+1)
		}
		lines = append(lines, "Phân loại: "+variant.Name+" (SKU "+variant.SKU+")")
		attributes["sku"+suffix] = variant.SKU
		if variant.PriceProvided {
			if price, err := decimal.NewFromString(variant.Price); err == nil {
				attributes["price"+suffix] = price.String()
			}
		}
		if variant.Stock != nil {
			attributes["stock"+suffix] = strconv.FormatInt(*variant.Stock, 10)
		}
	}
	if len(lines) == 0 {
		lines = append(lines, product.Name)
	}
	body := []rune(strings.Join(lines, "\n"))
	if len(body) > types.KnowledgeDocumentMaxBody {
		body = body[:types.KnowledgeDocumentMaxBody]
	}
	return types.KnowledgeSource{Revision: product.Revision, Upsert: &types.KnowledgeUpsert{SourceType: types.KnowledgeSourceProduct,
		SourceID: product.ID, Revision: product.Revision, Title: product.Name, Body: string(body), Attributes: attributes}}
}

// scriptKnowledge ingests only an active script's active version; version_id lets the AI service keep using a
// chunk only while that version is in the run's active script snapshot.
func scriptKnowledge(script *types.KnowledgeScript) types.KnowledgeSource {
	if script == nil {
		return types.KnowledgeSource{Revision: knowledgeGoneRevision}
	}
	if script.Status != types.ScriptStatusActive || script.ActiveVersionID == "" || strings.TrimSpace(script.Answer) == "" {
		return types.KnowledgeSource{Revision: script.Revision}
	}
	body := "Trả lời: " + script.Answer
	if len(script.Triggers) > 0 {
		body = "Khách hỏi: " + strings.Join(script.Triggers, "; ") + "\n" + body
	}
	return types.KnowledgeSource{Revision: script.Revision, Upsert: &types.KnowledgeUpsert{SourceType: types.KnowledgeSourceScript,
		SourceID: script.ID, Revision: script.Revision, Title: script.Name, Body: body, Attributes: map[string]string{"version_id": script.ActiveVersionID}}}
}

func documentKnowledge(document *types.KnowledgeDocument) types.KnowledgeSource {
	if document == nil {
		return types.KnowledgeSource{Revision: knowledgeGoneRevision}
	}
	if document.DeletedAt != nil {
		return types.KnowledgeSource{Revision: document.Version}
	}
	return types.KnowledgeSource{Revision: document.Version, Upsert: &types.KnowledgeUpsert{SourceType: types.KnowledgeSourceDocument,
		SourceID: document.ID, Revision: document.Version, Title: document.Title, Body: document.Body, Attributes: map[string]string{}}}
}
