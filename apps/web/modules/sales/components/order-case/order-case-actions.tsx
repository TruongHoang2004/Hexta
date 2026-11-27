"use client";

import { useState } from "react";
import { Textarea } from "@hexta/ui";
import { ConfirmDialog, Field, buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OrderCase } from "../../types";

interface Props {
  value: OrderCase;
  busy: boolean;
  canEdit: boolean;
  isOwner: boolean;
  /** Local quantity errors block saving before the API refuses it. */
  invalid: boolean;
  /** Finalizing a return needs saved returned lines. */
  canFinalize: boolean;
  onSave: () => void;
  onSendBack: (note: string) => Promise<void> | void;
  onFinalize: (note: string) => Promise<void> | void;
}

type Dialog = "sendBack" | "finalize" | null;

/** Handler progress save plus the owner-only send back and finalize decisions, each confirmed in a dialog. */
export function OrderCaseActions({ value, busy, canEdit, isOwner, invalid, canFinalize, onSave, onSendBack, onFinalize }: Props) {
  const t = useT();
  const [dialog, setDialog] = useState<Dialog>(null);
  const [note, setNote] = useState("");
  function show(next: Dialog) {
    setNote("");
    setDialog(next);
  }
  async function decide() {
    const action = dialog === "sendBack" ? onSendBack : onFinalize;
    setDialog(null);
    await action(note.trim());
  }
  if (value.status === "finalized") return null;
  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        {canEdit && (
          <button type="button" className={buttonClass} disabled={busy || invalid} onClick={onSave}>
            {busy ? t("orderCase.saving") : t("orderCase.save")}
          </button>
        )}
        {isOwner && value.status === "open" && (
          <button type="button" className={secondaryClass} disabled={busy} onClick={() => show("sendBack")}>
            {t("orderCase.sendBack")}
          </button>
        )}
        {isOwner && (
          <button type="button" className={secondaryClass} disabled={busy || !canFinalize} onClick={() => show("finalize")}>
            {t("orderCase.finalize")}
          </button>
        )}
      </div>
      {isOwner && !canFinalize && <p className="text-xs text-muted">{t("orderCase.finalizeNeedsLines")}</p>}
      <ConfirmDialog
        open={dialog !== null}
        title={dialog === "sendBack" ? t("orderCase.sendBackTitle") : t("orderCase.finalizeTitle")}
        confirmLabel={dialog === "sendBack" ? t("orderCase.sendBack") : t("orderCase.finalize")}
        cancelLabel={t("common.cancel")}
        confirmDisabled={dialog === "sendBack" && note.trim() === ""}
        busy={busy}
        onCancel={() => setDialog(null)}
        onConfirm={() => void decide()}
      >
        <p>{dialog === "sendBack" ? t("orderCase.sendBackHint") : t("orderCase.finalizeBody")}</p>
        <Field title={dialog === "sendBack" ? t("orderCase.sendBackReason") : t("orderCase.finalizeNote")}>
          <Textarea
            required={dialog === "sendBack"}
            maxLength={5000}
            value={note}
            onChange={(event) => setNote(event.target.value)}
          />
        </Field>
      </ConfirmDialog>
    </div>
  );
}
