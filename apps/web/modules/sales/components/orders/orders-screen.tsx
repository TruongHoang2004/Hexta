"use client";

import { Suspense, useCallback, useRef, useState } from "react";
import { Notice, buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { useOrderList } from "../../hooks/use-order-list";
import { useOrderLookups } from "../../hooks/use-order-lookups";
import { sales } from "../../services/sales";
import type { Order } from "../../types";
import { OrderDeepLink } from "./order-deep-link";
import { OrderEditor } from "./order-editor";
import { OrderList } from "./order-list";
import { OrderSessionLink } from "./order-session-link";

/** Order list beside the draft editor. Guards unsaved edits when switching orders. */
export function OrdersScreen() {
  const { tenant } = useWorkspace();
  const t = useT();
  const tenantID = tenant?.id ?? "";
  const [revision, setRevision] = useState(0);
  const [selected, setSelected] = useState<Order | null>(null);
  const [editorKey, setEditorKey] = useState(0);
  const [dirty, setDirty] = useState(false);
  const [busy, setBusy] = useState(false);
  const [success, setSuccess] = useState("");
  const [linkError, setLinkError] = useState("");
  // Open conversation session from the Desk link; the next new draft is linked to it.
  const [sessionTurn, setSessionTurn] = useState("");
  const errorText = useErrorText();
  // Bumped by every open so a slow linked-order load never replaces a later choice.
  const openRequest = useRef(0);
  const list = useOrderList(tenantID, revision);
  const lookups = useOrderLookups(tenantID, revision);
  const reload = () => setRevision((value) => value + 1);

  function saved(row: Order) {
    if (row.messaging_turn_id) setSessionTurn("");
    setDirty(false);
    setSelected(row);
    setEditorKey((value) => value + 1);
    reload();
    setSuccess(row.status === "draft" ? t("orders.savedDraft") : t("orders.confirmedOrder"));
  }
  function open(row: Order | null) {
    if (dirty && !window.confirm(t("orders.discardOpen"))) return;
    openRequest.current += 1;
    setDirty(false);
    setSelected(row);
    setEditorKey((value) => value + 1);
    setSuccess("");
    setLinkError("");
  }
  // A linked order (from the Overview) is loaded by ID because it may not be on the current list page.
  const openLinked = useCallback(
    (id: string) => {
      if (dirty && !window.confirm(t("orders.discardOpen"))) return;
      const request = ++openRequest.current;
      sales
        .order(tenantID, id)
        .then((row) => {
          if (request !== openRequest.current) return;
          setDirty(false);
          setSelected(row);
          setEditorKey((value) => value + 1);
          setLinkError("");
        })
        .catch((err: unknown) => {
          if (request === openRequest.current) setLinkError(errorText(err));
        });
    },
    [tenantID, dirty, t, errorText],
  );

  return (
    <div className="space-y-5">
      {tenantID && (
        <Suspense fallback={null}>
          <OrderDeepLink onOpen={openLinked} />
          <OrderSessionLink onLink={setSessionTurn} />
        </Suspense>
      )}
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("orders.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("orders.subtitle")}</p>
        </div>
        <div className="flex gap-2">
          <button type="button" className={secondaryClass} disabled={busy || list.loading} onClick={reload}>
            {t("orders.refresh")}
          </button>
          <button type="button" className={buttonClass} disabled={busy} onClick={() => open(null)}>
            {t("orders.newDraft")}
          </button>
        </div>
      </div>
      <Notice error={linkError || list.error || lookups.error} success={success} />
      {sessionTurn && !selected && (
        <p role="status" className="flex flex-wrap items-center gap-3 rounded-lg border border-border bg-surface p-3 text-sm">
          {t("orders.linkedSession")}
          <button type="button" className={secondaryClass} disabled={busy} onClick={() => setSessionTurn("")}>
            {t("orders.unlinkSession")}
          </button>
        </p>
      )}
      <div className="flex flex-wrap items-start gap-5">
        <OrderList list={list} customers={lookups.customers} selectedID={selected?.id} busy={busy} onOpen={open} />
        <div className="min-w-0 flex-[999_1_520px]">
          {lookups.loading && !lookups.members.length ? (
            <p role="status">{t("orders.loadingForm")}</p>
          ) : lookups.error ? (
            <div className="rounded-xl border border-border p-5">
              <p className="text-sm">{t("orders.lookupFailed")}</p>
              <button type="button" className={`${secondaryClass} mt-3`} onClick={reload}>
                {t("orders.retry")}
              </button>
            </div>
          ) : (
            <OrderEditor
              key={editorKey}
              order={selected}
              customers={lookups.customers}
              products={lookups.products}
              members={lookups.members}
              busy={busy}
              setBusy={setBusy}
              dirty={dirty}
              setDirty={setDirty}
              onSaved={saved}
              sessionTurnID={selected ? undefined : sessionTurn || undefined}
            />
          )}
        </div>
      </div>
    </div>
  );
}
