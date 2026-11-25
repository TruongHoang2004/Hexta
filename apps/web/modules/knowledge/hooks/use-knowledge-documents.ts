"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { knowledgeDocuments } from "../services/knowledge-documents";
import type { KnowledgeDocument } from "../types";

/** Paged knowledge documents, newest change first. Reloads when `revision` changes. */
export function useKnowledgeDocuments(tenantID: string, revision: number) {
  const errorText = useErrorText();
  const [page, setPage] = useState(1);
  const [rows, setRows] = useState<KnowledgeDocument[]>([]);
  const [total, setTotal] = useState(0);
  const [canManage, setCanManage] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!tenantID) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const data = await knowledgeDocuments.list(tenantID, page, abort.signal);
        if (abort.signal.aborted) return;
        setRows(data.items);
        setTotal(data.total);
        setCanManage(data.can_manage);
      } catch (err) {
        if (!abort.signal.aborted) setError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenantID, page, revision, errorText]);

  return { page, setPage, rows, total, canManage, loading, error };
}
