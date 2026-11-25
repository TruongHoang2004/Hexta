"use client";

import Link from "next/link";
import type { TenantMember } from "@hexta/sdk";
import { ConflictBanner, Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import { useOrderEditor, type OrderEditorOptions } from "../../hooks/use-order-editor";
import type { Customer, Product } from "../../types";
import { orderChecks } from "../../utils";
import { OrderCasePanel } from "../order-case/order-case-panel";
import { OrderChecklist } from "./order-checklist";
import { OrderDraftForm } from "./order-draft-form";
import { OrderEditorHeader } from "./order-editor-header";
import { OrderSavedSummary } from "./order-saved-summary";

interface Props extends Omit<OrderEditorOptions, "tenantID"> {
  customers: Customer[];
  products: Product[];
  members: TenantMember[];
}

/** Edit, claim and confirm one draft order; confirmed or others' drafts are read-only. Confirmed orders show their case panel. */
export function OrderEditor({ customers, products, members, ...options }: Props) {
  const { tenant, isOwner } = useWorkspace();
  const userID = useAuthStore((state) => state.user?.id);
  const t = useT();
  const { order, busy, dirty } = options;
  const editor = useOrderEditor({ ...options, tenantID: tenant?.id ?? "" });
  const readonly = order
    ? order.status !== "draft" || (!isOwner && !!order.employee_id && order.employee_id !== userID)
    : false;
  const checks = orderChecks(t, {
    order,
    dirty,
    userID,
    customer: customers.find((row) => row.id === editor.fields.customer),
    fields: editor.fields,
  });
  const claimable = !!order && order.status === "draft" && !order.employee_id && !!userID && !readonly;
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <OrderEditorHeader order={order} />
      {editor.conflict ? (
        <ConflictBanner
          message={editor.error}
          actionLabel={t("orders.editor.conflictReload")}
          onReload={() => void editor.reload(false)}
        />
      ) : (
        <Notice error={editor.error} />
      )}
      {readonly && (
        <p className="rounded-lg bg-background p-3 text-sm text-muted">
          {order?.status !== "draft" ? t("orders.editor.readonlyConfirmed") : t("orders.editor.readonlyOther")}
        </p>
      )}
      {order?.status === "confirmed" && (
        <Link href={`/tenants/${tenant?.id}/work`} className="inline-block text-sm text-primary underline">
          {t("orders.editor.linkedWork")}
        </Link>
      )}
      <div className="flex flex-wrap items-start gap-5">
        <OrderDraftForm
          editor={editor}
          order={order}
          customers={customers}
          products={products}
          members={members}
          busy={busy}
          readonly={readonly}
          claimable={claimable}
        />
        {(!order || order.status === "draft") && (
          <OrderChecklist
            checks={checks}
            canConfirm={!!order}
            disabled={busy || readonly}
            onConfirm={() => void editor.confirm()}
          />
        )}
      </div>
      {order && (
        <OrderSavedSummary order={order} dirty={dirty} busy={busy} onReload={() => void editor.reload(true)} />
      )}
      {order?.status === "confirmed" && (
        <OrderCasePanel
          order={order}
          products={products}
          members={members}
          dirty={dirty}
          setDirty={options.setDirty}
        />
      )}
    </section>
  );
}
