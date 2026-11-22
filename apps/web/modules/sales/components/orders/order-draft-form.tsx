"use client";

import { Checkbox, Textarea } from "@hexta/ui";
import Link from "next/link";
import type { TenantMember } from "@hexta/sdk";
import { Field, buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import type { useOrderEditor } from "../../hooks/use-order-editor";
import type { Customer, Order, Product } from "../../types";
import { OrderLinesField } from "./order-lines-field";
import { OrderPartyFields } from "./order-party-fields";

interface Props {
  editor: ReturnType<typeof useOrderEditor>;
  order: Order | null;
  customers: Customer[];
  products: Product[];
  members: TenantMember[];
  busy: boolean;
  readonly: boolean;
  claimable: boolean;
}

export function OrderDraftForm({ editor, order, customers, products, members, busy, readonly, claimable }: Props) {
  const { tenant } = useWorkspace();
  const userID = useAuthStore((state) => state.user?.id);
  const t = useT();
  const { fields, update } = editor;
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        void editor.persist(fields.employee);
      }}
      className="min-w-0 flex-[999_1_420px] space-y-5"
    >
      <fieldset disabled={busy || readonly} className="space-y-5">
        <OrderPartyFields
          customer={fields.customer}
          employee={fields.employee}
          onChange={update}
          customers={customers}
          members={members}
          assigned={!!order?.employee_id}
        />
        <div className="flex flex-wrap gap-4 text-xs">
          <Link href={`/tenants/${tenant?.id}/customers`} className="text-primary underline">
            {t("orders.editor.manageCustomers")}
          </Link>
          <Link href={`/tenants/${tenant?.id}/products`} className="text-primary underline">
            {t("orders.editor.manageProducts")}
          </Link>
          <Link href={`/tenants/${tenant?.id}/settings/members`} className="text-primary underline">
            {t("orders.editor.manageTeam")}
          </Link>
        </div>
        <OrderLinesField
          lines={fields.lines}
          products={products}
          onAdd={editor.addLine}
          onUpdate={editor.updateLine}
          onRemove={editor.removeLine}
        />
        <label className="flex min-h-11 items-center gap-2 text-sm">
          <Checkbox checked={fields.delivery} onCheckedChange={(checked) => update({ delivery: checked === true })} />
          {t("orders.editor.delivery")}
        </label>
        {fields.delivery && (
          <Field title={t("orders.editor.address")}>
            <Textarea
              maxLength={5000}
              rows={2}
              value={fields.address}
              onChange={(e) => update({ address: e.target.value })}
            />
          </Field>
        )}
        <Field title={t("orders.editor.notes")}>
          <Textarea
            rows={3}
            maxLength={10000}
            value={fields.notes}
            onChange={(e) => update({ notes: e.target.value })}
          />
        </Field>
        {!readonly && (
          <div className="flex flex-wrap gap-2">
            <button type="submit" className={buttonClass}>
              {busy ? t("orders.editor.saving") : t("orders.editor.saveDraft")}
            </button>
            {claimable && (
              <button type="button" className={secondaryClass} onClick={() => userID && void editor.persist(userID)}>
                {t("orders.editor.claim")}
              </button>
            )}
          </div>
        )}
      </fieldset>
    </form>
  );
}
