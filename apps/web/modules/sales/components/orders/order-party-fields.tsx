"use client";

import type { TenantMember } from "@hexta/sdk";
import { Field, inputClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import { useAuthStore } from "@/store/useAuthStore";
import type { Customer } from "../../types";

interface Props {
  customer: string;
  employee: string;
  onChange: (patch: { customer?: string; employee?: string }) => void;
  customers: Customer[];
  members: TenantMember[];
  /** The saved order already has an assignee; members may not unassign it. */
  assigned: boolean;
}

/** Customer and assignee pickers; members can only assign themselves. */
export function OrderPartyFields({ customer, employee, onChange, customers, members, assigned }: Props) {
  const { isOwner } = useWorkspace();
  const userID = useAuthStore((state) => state.user?.id);
  const t = useT();
  const memberName = useMemberName(members);
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Field title={t("orders.editor.customer")}>
        <select className={inputClass} value={customer} onChange={(e) => onChange({ customer: e.target.value })}>
          <option value="">{t("orders.editor.selectLater")}</option>
          {customer && !customers.some((row) => row.id === customer) && (
            <option value={customer}>{t("orders.unavailableCustomer")}</option>
          )}
          {customers.map((row) => (
            <option key={row.id} value={row.id}>
              {row.full_name}
              {row.phone ? ` · ${row.phone}` : ""}
            </option>
          ))}
        </select>
      </Field>
      <Field title={t("orders.editor.assignee")}>
        <select className={inputClass} value={employee} onChange={(e) => onChange({ employee: e.target.value })}>
          <option value="" disabled={!isOwner && assigned}>
            {t("orders.editor.unassigned")}
          </option>
          {employee && !members.some((row) => row.user_id === employee) && (
            <option value={employee}>{t("orders.editor.formerMember")}</option>
          )}
          {members.map((row) => (
            <option key={row.id} value={row.user_id} disabled={!isOwner && row.user_id !== userID}>
              {memberName(row.user_id)}
            </option>
          ))}
        </select>
      </Field>
    </div>
  );
}
