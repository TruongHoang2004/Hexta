import { sdk } from "@/lib/sdk";
import type { KnowledgeDocument, KnowledgeDocumentInput, KnowledgeDocumentPage } from "../types";

const base = (tenant: string) => `/api/v1/tenants/${tenant}/knowledge-documents`;

/** Knowledge document endpoints; every member reads, only the owner writes, and the API authorizes every call. */
export const knowledgeDocuments = {
  async list(tenant: string, page: number, signal?: AbortSignal): Promise<KnowledgeDocumentPage> {
    const res = await sdk.client.get<{ data: KnowledgeDocumentPage }>(base(tenant), {
      params: { page, size: 20 },
      signal,
    });
    return res.data.data;
  },
  async create(tenant: string, input: KnowledgeDocumentInput): Promise<KnowledgeDocument> {
    const res = await sdk.client.post<{ data: KnowledgeDocument }>(base(tenant), input);
    return res.data.data;
  },
  /** expected_version is the version last read; a stale version is rejected with 409. */
  async update(tenant: string, document: KnowledgeDocument, input: KnowledgeDocumentInput): Promise<KnowledgeDocument> {
    const res = await sdk.client.put<{ data: KnowledgeDocument }>(`${base(tenant)}/${document.id}`, {
      ...input,
      expected_version: document.version,
    });
    return res.data.data;
  },
  async remove(tenant: string, document: KnowledgeDocument): Promise<void> {
    await sdk.client.delete(`${base(tenant)}/${document.id}`, { params: { expected_version: document.version } });
  },
};
