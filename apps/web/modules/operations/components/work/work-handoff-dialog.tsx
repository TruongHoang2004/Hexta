"use client";

import { useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { ConfirmDialog, Field, inputClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import { useAuthStore } from "@/store/useAuthStore";
import type { WorkItem } from "../../types";

interface Props {
  open: boolean;
  item: WorkItem;
  members: TenantMember[];
  busy: boolean;
  onCancel: () => void;
  onConfirm: (recipientID: string) => Promise<void>;
}

/** Pick who receives the next step; explains whether they must accept first. */
export function WorkHandoffDialog({ open, item, members, busy, onCancel, onConfirm }: Props) {
  const t = useT();
  const person = useMemberName(members);
  const userID = useAuthStore((state) => state.user?.id);
  const [recipient, setRecipient] = useState("");
  const currentStep = item.workflow.steps[item.current_step];
  const nextStep = item.workflow.steps[item.current_step + 1];
  return (
    <ConfirmDialog
      open={open}
      title={t("work.handoffDialog.title")}
      confirmLabel={t("work.handoffDialog.confirm")}
      cancelLabel={t("common.cancel")}
      busy={busy}
      confirmDisabled={!recipient}
      onCancel={onCancel}
      onConfirm={() => onConfirm(recipient)}
    >
      {nextStep && <p className="font-medium">{t("work.handoffDialog.next", { name: nextStep.name })}</p>}
      <Field title={t("work.handoffDialog.recipient")}>
        <select className={inputClass} value={recipient} onChange={(e) => setRecipient(e.target.value)}>
          <option value="">{t("work.handoffDialog.selectRecipient")}</option>
          {members
            .filter((row) => row.user_id !== userID)
            .map((row) => (
              <option key={row.id} value={row.user_id}>
                {person(row.user_id)}
              </option>
            ))}
        </select>
      </Field>
      <p className="rounded-lg bg-background p-3">
        {currentStep?.handoff_mode === "acceptance" ? t("work.handoffDialog.acceptance") : t("work.handoffDialog.direct")}
      </p>
    </ConfirmDialog>
  );
}
