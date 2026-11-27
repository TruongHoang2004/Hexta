"use client";

import { Textarea } from "@hexta/ui";
import type { TenantMember } from "@hexta/sdk";
import { ConflictBanner, Field, Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import { useAuthStore } from "@/store/useAuthStore";
import { useOrderCase } from "../../hooks/use-order-case";
import type { Order, Product } from "../../types";
import { variantOptions } from "../../utils";
import { OrderCaseActions } from "./order-case-actions";
import { OrderCaseHeader } from "./order-case-header";
import { OrderCaseLines } from "./order-case-lines";
import { OrderCaseOpenActions } from "./order-case-open-actions";
import { OrderCaseTimeline } from "./order-case-timeline";

interface Props {
  order: Order;
  products: Product[];
  members: TenantMember[];
  dirty: boolean;
  setDirty: (dirty: boolean) => void;
}

/**
 * Cancellation/return case of a confirmed order (S-10 additions). The handler or owner records progress and
 * returned quantities; only the owner sends back or finalizes; a finalized case is read-only. The API decides.
 */
export function OrderCasePanel({ order, products, members, dirty, setDirty }: Props) {
  const t = useT();
  const { tenant, isOwner } = useWorkspace();
  const userID = useAuthStore((state) => state.user?.id);
  const person = useMemberName(members);
  const state = useOrderCase(tenant?.id ?? "", order, setDirty);
  const variants = variantOptions(products);
  const itemLabel = (variantID: string, itemID: string) =>
    variants.find((row) => row.id === variantID)?.label ?? t("orderCase.lines.unknownProduct", { id: itemID.slice(0, 8) });
  const value = state.value;
  const canEdit = !!value && value.status !== "finalized" && (isOwner || value.handler_id === userID);
  const invalid = Object.keys(state.quantityErrors).length > 0;

  return (
    <section aria-labelledby="order-case-title" className="space-y-4 border-t border-border pt-4">
      <h2 id="order-case-title" className="sr-only">
        {value ? t(`orderCase.title.${value.type}`) : t("orderCase.openReturn")}
      </h2>
      <p className="rounded-lg bg-background p-3 text-xs text-muted">{t("orderCase.policy")}</p>
      {state.conflict ? (
        <ConflictBanner message={state.error} actionLabel={t("orderCase.conflictReload")} onReload={state.reload} />
      ) : (
        <Notice error={state.error} success={state.success} />
      )}
      {state.loading ? (
        <p role="status" className="text-sm">
          {t("orderCase.loading")}
        </p>
      ) : !value ? (
        !state.error && <OrderCaseOpenActions busy={state.busy} onOpen={(type) => void state.open(type)} />
      ) : (
        <div className="space-y-4">
          <OrderCaseHeader
            value={value}
            members={members}
            person={person}
            canReassign={isOwner && value.status !== "finalized"}
            handler={state.form.handler}
            onHandler={(handler) => state.edit({ handler })}
          />
          {value.status !== "finalized" && !canEdit && (
            <p className="rounded-lg bg-background p-3 text-sm text-muted">{t("orderCase.readonlyOther")}</p>
          )}
          {value.status === "finalized" && (
            <p className="text-sm text-muted">{t("orderCase.readonlyFinalized")}</p>
          )}
          {value.type === "return" ? (
            <OrderCaseLines
              order={order}
              quantities={state.form.quantities}
              errors={state.quantityErrors}
              readonly={!canEdit || state.busy}
              itemLabel={itemLabel}
              onChange={(itemID, quantity) => state.edit({ quantities: { ...state.form.quantities, [itemID]: quantity } })}
            />
          ) : (
            <p className="text-sm text-muted">{t("orderCase.cancellationNote")}</p>
          )}
          <Field title={t("orderCase.note")}>
            <Textarea
              maxLength={5000}
              disabled={!canEdit || state.busy}
              value={state.form.note}
              onChange={(event) => state.edit({ note: event.target.value })}
            />
          </Field>
          {dirty && canEdit && (
            <p role="status" className="text-sm">
              {t("orderCase.unsaved")}
            </p>
          )}
          <OrderCaseActions
            value={value}
            busy={state.busy}
            canEdit={canEdit}
            isOwner={isOwner}
            invalid={invalid}
            canFinalize={!dirty && (value.type === "cancellation" || value.lines.length > 0)}
            onSave={() => void state.save()}
            onSendBack={state.sendBack}
            onFinalize={state.finalize}
          />
          <OrderCaseTimeline events={value.events} person={person} />
        </div>
      )}
    </section>
  );
}
