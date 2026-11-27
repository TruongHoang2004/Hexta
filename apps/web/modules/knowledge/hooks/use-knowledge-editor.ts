"use client";

import { useReducer } from "react";
import { isConflict, useErrorText } from "@/lib/errors";
import { knowledgeDocuments } from "../services/knowledge-documents";
import type { KnowledgeDocument, KnowledgeDocumentInput } from "../types";

export type KnowledgeNotice = "created" | "updated" | "deleted" | null;

interface State {
  /** null: no form open; otherwise the document being edited, or "new". */
  editing: KnowledgeDocument | "new" | null;
  form: KnowledgeDocumentInput;
  /** Document waiting for delete confirmation. */
  deleting: KnowledgeDocument | null;
  busy: boolean;
  error: string;
  conflict: boolean;
  notice: KnowledgeNotice;
}

type Action =
  | { type: "open"; document: KnowledgeDocument | "new" }
  | { type: "close" }
  | { type: "change"; patch: Partial<KnowledgeDocumentInput> }
  | { type: "askDelete"; document: KnowledgeDocument | null }
  | { type: "start" }
  | { type: "done"; notice: KnowledgeNotice }
  | { type: "fail"; error: string; conflict: boolean }
  | { type: "reset" };

const empty: KnowledgeDocumentInput = { title: "", body: "" };

const initial: State = { editing: null, form: empty, deleting: null, busy: false, error: "", conflict: false, notice: null };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case "open":
      return {
        ...state,
        editing: action.document,
        form: action.document === "new" ? empty : { title: action.document.title, body: action.document.body },
        error: "",
        conflict: false,
        notice: null,
      };
    case "close":
      return { ...state, editing: null, form: empty, error: "", conflict: false };
    case "change":
      return { ...state, form: { ...state.form, ...action.patch } };
    case "askDelete":
      return { ...state, deleting: action.document, error: "", conflict: false, notice: null };
    case "start":
      return { ...state, busy: true, error: "", conflict: false, notice: null };
    case "done":
      return { ...initial, notice: action.notice };
    case "fail":
      return { ...state, busy: false, deleting: null, error: action.error, conflict: action.conflict };
    case "reset":
      return initial;
  }
}

/**
 * Create, edit and delete for the owner. Every write sends the version last read, so a document changed meanwhile
 * is a conflict the owner resolves by reloading; `onSaved` reloads the list.
 */
export function useKnowledgeEditor(tenantID: string, onSaved: () => void) {
  const errorText = useErrorText();
  const [state, dispatch] = useReducer(reducer, initial);

  async function run(write: () => Promise<unknown>, notice: KnowledgeNotice) {
    if (state.busy) return;
    dispatch({ type: "start" });
    try {
      await write();
      dispatch({ type: "done", notice });
      onSaved();
    } catch (err) {
      const conflict = isConflict(err);
      dispatch({ type: "fail", error: conflict ? "" : errorText(err), conflict });
    }
  }

  function save() {
    const editing = state.editing;
    if (!editing) return;
    const input = { title: state.form.title.trim(), body: state.form.body.trim() };
    void run(
      () => (editing === "new" ? knowledgeDocuments.create(tenantID, input) : knowledgeDocuments.update(tenantID, editing, input)),
      editing === "new" ? "created" : "updated",
    );
  }

  function confirmDelete() {
    const document = state.deleting;
    if (!document) return;
    void run(() => knowledgeDocuments.remove(tenantID, document), "deleted");
  }

  return {
    ...state,
    valid: state.form.title.trim().length > 0 && state.form.body.trim().length > 0,
    open: (document: KnowledgeDocument | "new") => dispatch({ type: "open", document }),
    close: () => dispatch({ type: "close" }),
    change: (patch: Partial<KnowledgeDocumentInput>) => dispatch({ type: "change", patch }),
    askDelete: (document: KnowledgeDocument | null) => dispatch({ type: "askDelete", document }),
    save,
    confirmDelete,
    /** After a conflict: drop the stale form and reload the list. */
    reload: () => {
      dispatch({ type: "reset" });
      onSaved();
    },
  };
}
