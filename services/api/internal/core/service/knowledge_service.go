package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

// KnowledgeService owns the tenant's business knowledge documents (FAQ, policies) the bot may answer from. Every
// active member reads them; only the tenant owner creates, edits or deletes them, with an optimistic version. Each
// change enqueues a durable knowledge sync job in the same transaction; the AI service is never called inline.
type KnowledgeService struct {
	store repository.KnowledgeDocumentStore
	now   func() time.Time
}

func NewKnowledgeService(store repository.KnowledgeDocumentStore) *KnowledgeService {
	return &KnowledgeService{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *KnowledgeService) within(ctx context.Context, tenant, actor string, owner bool, fn func(repository.KnowledgeDocumentStore, types.MessagingScope) *apperrors.Error) *apperrors.Error {
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	return s.store.WithinKnowledge(ctx, func(r repository.KnowledgeDocumentStore) *apperrors.Error {
		scope, err := r.Scope(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if !scope.Active || !scope.Member {
			return apperrors.ErrForbidden(ctx).SetDetail("active tenant membership required")
		}
		if owner && !scope.Owner {
			return apperrors.ErrForbidden(ctx).SetDetail("only the tenant owner can change knowledge documents")
		}
		return fn(r, scope)
	})
}

func normalizeDocument(ctx context.Context, input *types.KnowledgeDocumentInput) *apperrors.Error {
	input.Title, input.Body = strings.TrimSpace(input.Title), strings.TrimSpace(input.Body)
	if !utf8.ValidString(input.Title) || !utf8.ValidString(input.Body) {
		return apperrors.ErrBadRequest(ctx).SetDetail("title and body must be valid UTF-8 text")
	}
	if n := utf8.RuneCountInString(input.Title); n < 1 || n > types.KnowledgeDocumentMaxTitle {
		return apperrors.ErrBadRequest(ctx).SetDetail("title must be 1-200 characters")
	}
	if n := utf8.RuneCountInString(input.Body); n < 1 || n > types.KnowledgeDocumentMaxBody {
		return apperrors.ErrBadRequest(ctx).SetDetail("body must be 1-20000 characters")
	}
	return nil
}

func documentID(ctx context.Context, id string) *apperrors.Error {
	if _, err := uuid.Parse(id); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("invalid knowledge document ID")
	}
	return nil
}

// List returns live documents, newest change first, with their sync state and whether the caller may manage them.
func (s *KnowledgeService) List(ctx context.Context, tenant, actor string, page, size int) (types.KnowledgeDocumentPage, *apperrors.Error) {
	var result types.KnowledgeDocumentPage
	err := s.within(ctx, tenant, actor, false, func(r repository.KnowledgeDocumentStore, scope types.MessagingScope) *apperrors.Error {
		limit, offset := pageBounds(page, size)
		items, total, err := r.Documents(ctx, tenant, limit, offset)
		if err != nil {
			return err
		}
		result = types.KnowledgeDocumentPage{Items: items, Total: total, CanManage: scope.Owner}
		return nil
	})
	return result, err
}

// Create adds a document as the tenant owner.
func (s *KnowledgeService) Create(ctx context.Context, tenant, actor string, input types.KnowledgeDocumentInput) (*types.KnowledgeDocument, *apperrors.Error) {
	if err := normalizeDocument(ctx, &input); err != nil {
		return nil, err
	}
	var result *types.KnowledgeDocument
	err := s.within(ctx, tenant, actor, true, func(r repository.KnowledgeDocumentStore, _ types.MessagingScope) *apperrors.Error {
		now := s.now()
		document := types.KnowledgeDocument{ID: uuid.NewString(), TenantID: tenant, Title: input.Title, Body: input.Body, Version: 1,
			CreatedBy: actor, UpdatedBy: actor, CreatedAt: now, UpdatedAt: now}
		if err := r.InsertDocument(ctx, document); err != nil {
			return err
		}
		if err := r.EnqueueKnowledgeSync(ctx, tenant, types.KnowledgeSourceDocument, document.ID, now); err != nil {
			return err
		}
		document.SyncState = types.KnowledgeSyncPending
		result = &document
		return r.Audit(ctx, tenant, actor, "knowledge.document_created", "knowledge_document", document.ID, map[string]any{"version": document.Version})
	})
	return result, err
}

// Update rewrites a document as the tenant owner; a stale ExpectedVersion is a conflict.
func (s *KnowledgeService) Update(ctx context.Context, tenant, actor, id string, input types.KnowledgeDocumentInput) (*types.KnowledgeDocument, *apperrors.Error) {
	if err := documentID(ctx, id); err != nil {
		return nil, err
	}
	if err := normalizeDocument(ctx, &input); err != nil {
		return nil, err
	}
	var result *types.KnowledgeDocument
	err := s.within(ctx, tenant, actor, true, func(r repository.KnowledgeDocumentStore, _ types.MessagingScope) *apperrors.Error {
		document, err := r.Document(ctx, tenant, id, true)
		if err != nil {
			return err
		}
		if document.Version != input.ExpectedVersion {
			return apperrors.ErrConflict(ctx, "knowledge document", "changed; refresh before retrying")
		}
		if document.Title == input.Title && document.Body == input.Body {
			result = document
			return nil
		}
		now := s.now()
		document.Title, document.Body, document.Version, document.UpdatedBy, document.UpdatedAt = input.Title, input.Body, document.Version+1, actor, now
		if err := r.SaveDocument(ctx, *document); err != nil {
			return err
		}
		if err := r.EnqueueKnowledgeSync(ctx, tenant, types.KnowledgeSourceDocument, id, now); err != nil {
			return err
		}
		document.SyncState, document.SyncError = types.KnowledgeSyncPending, ""
		result = document
		return r.Audit(ctx, tenant, actor, "knowledge.document_updated", "knowledge_document", id, map[string]any{"version": document.Version})
	})
	return result, err
}

// Delete removes a document as the tenant owner. The row stays as a tombstone with a newer version so the AI
// knowledge removal carries a revision newer than any earlier upsert.
func (s *KnowledgeService) Delete(ctx context.Context, tenant, actor, id string, expected int64) *apperrors.Error {
	if err := documentID(ctx, id); err != nil {
		return err
	}
	return s.within(ctx, tenant, actor, true, func(r repository.KnowledgeDocumentStore, _ types.MessagingScope) *apperrors.Error {
		document, err := r.Document(ctx, tenant, id, true)
		if err != nil {
			return err
		}
		if document.Version != expected {
			return apperrors.ErrConflict(ctx, "knowledge document", "changed; refresh before retrying")
		}
		now := s.now()
		document.Version, document.UpdatedBy, document.UpdatedAt, document.DeletedAt = document.Version+1, actor, now, &now
		if err := r.SaveDocument(ctx, *document); err != nil {
			return err
		}
		if err := r.EnqueueKnowledgeSync(ctx, tenant, types.KnowledgeSourceDocument, id, now); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "knowledge.document_deleted", "knowledge_document", id, map[string]any{"version": document.Version})
	})
}
