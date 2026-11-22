"use client";

import { useState } from "react";
import { ConfirmDialog } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { handoffReasons, type HandoffReason } from "../../types";

/** Confirms handing the conversation to the waiting queue, with a reason. */
export function HandoffDialog({
  open,
  busy,
  onCancel,
  onConfirm,
}: {
  open: boolean;
  busy: boolean;
  onCancel: () => void;
  onConfirm: (reason: HandoffReason) => void;
}) {
  const t = useT();
  const [reason, setReason] = useState<HandoffReason>("staff_handoff");
  return (
    <ConfirmDialog
      open={open}
      title={t("desk.handoffDialog.title")}
      confirmLabel={t("desk.handoffDialog.confirm")}
      cancelLabel={t("common.cancel")}
      busy={busy}
      onCancel={onCancel}
      onConfirm={() => onConfirm(reason)}
    >
      <p>{t("desk.handoffDialog.body")}</p>
      <fieldset className="space-y-1">
        <legend className="mb-1 font-medium">{t("desk.handoffDialog.reasonLegend")}</legend>
        {handoffReasons.map((code) => (
          <label key={code} className="flex min-h-9 items-center gap-2">
            <input
              type="radio"
              name="handoff-reason"
              value={code}
              checked={reason === code}
              onChange={() => setReason(code)}
            />
            {t(`desk.reasons.${code}`)}
          </label>
        ))}
      </fieldset>
    </ConfirmDialog>
  );
}
