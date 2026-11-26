/** Mirrors the Go API knowledge document DTOs (services/api/internal/present/http/dto). */

/**
 * AI knowledge sync status: pending = waiting to be indexed for the bot, done = indexed, failed = rejected
 * (sync_error holds a code). Empty when the API has no sync record yet.
 */
export type KnowledgeSyncState = "pending" | "done" | "failed" | "";

/** A business document (FAQ, policy) the bot may answer from. */
export interface KnowledgeDocument {
  id: string;
  title: string;
  body: string;
  version: number;
  sync_state: KnowledgeSyncState;
  sync_error?: string;
  created_by: string;
  updated_by: string;
  created_at: string;
  updated_at: string;
}

export interface KnowledgeDocumentPage {
  items: KnowledgeDocument[];
  total: number;
  page: number;
  size: number;
  /** The caller is the tenant owner and may create, edit and delete documents. */
  can_manage: boolean;
}

export interface KnowledgeDocumentInput {
  title: string;
  body: string;
}

/** API limits; the server trims and enforces them again. */
export const knowledgeLimits = { title: 200, body: 20000 } as const;
