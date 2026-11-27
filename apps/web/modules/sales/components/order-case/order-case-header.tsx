"use client";

import type { TenantMember } from "@hexta/sdk";
import { Field, Picker, StatusChip, type ChipTone } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OrderCase, OrderCaseStatus } from "../../types";

const tones: Record<OrderCaseStatus, ChipTone> = { open: "warning", sent_back: "danger", finalized: "success" };

interface Props {
  value: OrderCase;
  members: TenantMember[];
  person: (id: string | null | undefined) => string;
  /** Owner may reassign the handler while the case is not finalized. */
  canReassign: boolean;
  handler: string;
  onHandler: (id: string) => void;
}

/** Case type, number, status, handler and the finalized stamp. */
export function OrderCaseHeader({ value, members, person, canReassign, handler, onHandler }: Props) {
  const t = useT();
  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <h3 className="font-semibold">
          {t(`orderCase.title.${value.type}`)} {t("orderCase.number", { id: value.id.slice(0, 8).toUpperCase() })}
        </h3>
        <StatusChip tone={tones[value.status]}>{t(`orderCase.status.${value.status}`)}</StatusChip>
      </div>
      {canReassign ? (
        <Field title={t("orderCase.handler")}>
          <Picker
            value={handler}
            onChange={onHandler}
            options={members.map((member) => ({ value: member.user_id, label: person(member.user_id) }))}
          />
        </Field>
      ) : (
        <p className="text-sm">
          {t("orderCase.handler")}: <strong>{person(value.handler_id)}</strong>
        </p>
      )}
      {value.status === "finalized" && (
        <p role="status" className="rounded-lg bg-background p-3 text-sm text-muted">
          {t("orderCase.finalizedBy", {
            name: person(value.finalized_by),
            time: value.finalized_at ? new Date(value.finalized_at).toLocaleString() : "",
          })}
        </p>
      )}
    </div>
  );
}
