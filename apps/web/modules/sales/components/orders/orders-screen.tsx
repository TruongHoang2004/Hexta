"use client";

import { useState } from "react";
import { Notice, buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useOrderList } from "../../hooks/use-order-list";
import { useOrderLookups } from "../../hooks/use-order-lookups";
import type { Order } from "../../types";
import { OrderEditor } from "./order-editor";
import { OrderList } from "./order-list";

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
  const list = useOrderList(tenantID, revision);
  const lookups = useOrderLookups(tenantID, revision);
  const reload = () => setRevision((value) => value + 1);

  function saved(row: Order) {
    setDirty(false);
    setSelected(row);
    setEditorKey((value) => value + 1);
    reload();
    setSuccess(row.status === "draft" ? t("orders.savedDraft") : t("orders.confirmedOrder"));
  }
  function open(row: Order | null) {
    if (dirty && !window.confirm(t("orders.discardOpen"))) return;
    setDirty(false);
    setSelected(row);
    setEditorKey((value) => value + 1);
    setSuccess("");
  }

  return (
    <div className="space-y-5">
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
      <Notice error={list.error || lookups.error} success={success} />
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
            />
          )}
        </div>
      </div>
    </div>
  );
}
