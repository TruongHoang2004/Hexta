"use client";

import { useEffect, useReducer } from "react";
import { SDKApiError } from "@hexta/sdk";
import { isConflict, useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { orderCases } from "../services/order-cases";
import type { Order, OrderCase, OrderCaseForm, OrderCaseType } from "../types";
import { caseFormFrom, caseUpdateInput, returnQuantityErrors } from "../utils";

interface State {
  loading: boolean;
  busy: boolean;
  value: OrderCase | null;
  form: OrderCaseForm;
  error: string;
  conflict: boolean;
  success: string;
  revision: number;
}

type Action =
  | { type: "loading" }
  | { type: "loaded"; value: OrderCase | null; success?: string }
  | { type: "busy" }
  | { type: "failed"; error: string; conflict: boolean }
  | { type: "edit"; patch: Partial<OrderCaseForm> }
  | { type: "reload" };

const emptyForm: OrderCaseForm = { note: "", handler: "", quantities: {} };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case "loading":
      return { ...state, loading: true, error: "", conflict: false };
    case "loaded":
      return {
        ...state,
        loading: false,
        busy: false,
        value: action.value,
        form: action.value ? caseFormFrom(action.value) : emptyForm,
        error: "",
        conflict: false,
        success: action.success ?? "",
      };
    case "busy":
      return { ...state, busy: true, error: "", conflict: false, success: "" };
    case "failed":
      return { ...state, loading: false, busy: false, error: action.error, conflict: action.conflict };
    case "edit":
      return { ...state, form: { ...state.form, ...action.patch }, success: "" };
    case "reload":
      return { ...state, revision: state.revision + 1 };
  }
}

/**
 * Cancellation/return case of one confirmed order: load, open, record progress and owner decisions.
 * `onDirty` lets the screen guard unsaved case edits like draft edits.
 */
export function useOrderCase(tenantID: string, order: Order, onDirty: (dirty: boolean) => void) {
  const t = useT();
  const errorText = useErrorText();
  const [state, dispatch] = useReducer(reducer, {
    loading: true,
    busy: false,
    value: null,
    form: emptyForm,
    error: "",
    conflict: false,
    success: "",
    revision: 0,
  });

  useEffect(() => {
    if (!tenantID) return;
    const abort = new AbortController();
    dispatch({ type: "loading" });
    orderCases
      .get(tenantID, order.id, abort.signal)
      .then((value) => {
        if (!abort.signal.aborted) dispatch({ type: "loaded", value });
      })
      .catch((err: unknown) => {
        if (!abort.signal.aborted) dispatch({ type: "failed", error: errorText(err), conflict: false });
      });
    return () => abort.abort();
  }, [tenantID, order.id, state.revision, errorText]);

  /** Runs one mutation; the button is disabled while busy, so a retry never duplicates it. */
  /** `ownerOnly` shows the owner-only finalize message instead of the generic forbidden text. */
  async function run(action: () => Promise<OrderCase>, success: string, ownerOnly = false) {
    if (state.busy) return;
    dispatch({ type: "busy" });
    try {
      dispatch({ type: "loaded", value: await action(), success });
      onDirty(false);
    } catch (err) {
      const forbidden = ownerOnly && err instanceof SDKApiError && err.status === 403;
      dispatch({ type: "failed", error: forbidden ? t("orderCase.ownerOnly") : errorText(err), conflict: isConflict(err) });
    }
  }

  const value = state.value;
  const quantityErrors = returnQuantityErrors(order, state.form);
  return {
    ...state,
    quantityErrors,
    edit(patch: Partial<OrderCaseForm>) {
      onDirty(true);
      dispatch({ type: "edit", patch });
    },
    reload() {
      onDirty(false);
      dispatch({ type: "reload" });
    },
    open(type: OrderCaseType) {
      return run(() => orderCases.open(tenantID, order.id, type), t("orderCase.opened"));
    },
    save() {
      if (!value || Object.keys(quantityErrors).length) return;
      return run(() => orderCases.update(tenantID, order.id, caseUpdateInput(value, order, state.form)), t("orderCase.saved"));
    },
    sendBack(note: string) {
      if (!value) return;
      return run(() => orderCases.sendBack(tenantID, order.id, value.version, note), t("orderCase.sentBackDone"));
    },
    finalize(note: string) {
      if (!value) return;
      return run(() => orderCases.finalize(tenantID, order.id, value.version, note), t("orderCase.finalizedDone"), true);
    },
  };
}
