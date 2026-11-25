// Package httpmap converts knowledge documents to HTTP DTOs.
package httpmap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

// Document maps a document; a job in progress (leased) is reported as pending.
func Document(value types.KnowledgeDocument) dto.KnowledgeDocument {
	state := value.SyncState
	if state == types.KnowledgeSyncLeased {
		state = types.KnowledgeSyncPending
	}
	return dto.KnowledgeDocument{ID: value.ID, Title: value.Title, Body: value.Body, Version: value.Version, SyncState: state,
		SyncError: value.SyncError, CreatedBy: value.CreatedBy, UpdatedBy: value.UpdatedBy, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}

func Documents(value types.KnowledgeDocumentPage, page, size int) dto.KnowledgeDocumentsResponse {
	result := dto.KnowledgeDocumentsResponse{Items: make([]dto.KnowledgeDocument, len(value.Items)), Total: value.Total, Page: page, Size: size, CanManage: value.CanManage}
	for i, item := range value.Items {
		result.Items[i] = Document(item)
	}
	return result
}
