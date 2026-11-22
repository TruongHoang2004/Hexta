"use client";

import type { TenantMember } from "@hexta/sdk";
import { Field, Picker, type PickerOption } from "@/components/primitives";
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

/** Customer and assignee pickers; members can only assign themselves. A disabled fieldset disables both. */
export function OrderPartyFields({ customer, employee, onChange, customers, members, assigned }: Props) {
  const { isOwner } = useWorkspace();
  const userID = useAuthStore((state) => state.user?.id);
  const t = useT();
  const memberName = useMemberName(members);
  const customerOptions: PickerOption[] = [
    { value: "", label: t("orders.editor.selectLater") },
    ...(customer && !customers.some((row) => row.id === customer)
      ? [{ value: customer, label: t("orders.unavailableCustomer") }]
      : []),
    ...customers.map((row) => ({ value: row.id, label: row.full_name, description: row.phone || row.email })),
  ];
  const memberOptions: PickerOption[] = [
    { value: "", label: t("orders.editor.unassigned"), disabled: !isOwner && assigned },
    ...(employee && !members.some((row) => row.user_id === employee)
      ? [{ value: employee, label: t("orders.editor.formerMember") }]
      : []),
    ...members.map((row) => ({
      value: row.user_id,
      label: memberName(row.user_id),
      disabled: !isOwner && row.user_id !== userID,
    })),
  ];
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Field title={t("orders.editor.customer")}>
        <Picker
          options={customerOptions}
          value={customer}
          onChange={(value) => onChange({ customer: value })}
        />
      </Field>
      <Field title={t("orders.editor.assignee")}>
        <Picker
          options={memberOptions}
          value={employee}
          onChange={(value) => onChange({ employee: value })}
        />
      </Field>
    </div>
  );
}
