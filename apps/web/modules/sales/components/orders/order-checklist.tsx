"use client";

import { ChecklistItem, buttonClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OrderCheck } from "../../types";

interface Props {
  checks: OrderCheck[];
  canConfirm: boolean;
  disabled: boolean;
  onConfirm: () => void;
}

/** Confirmation readiness; the confirm button explains itself via the live summary. */
export function OrderChecklist({ checks, canConfirm, disabled, onConfirm }: Props) {
  const t = useT();
  const failing = checks.filter((check) => !check.passed).length;
  return (
    <aside className="flex-[1_1_260px] space-y-3 rounded-lg border border-border bg-background p-4">
      <h3 className="font-semibold">{t("orders.checklist.title")}</h3>
      <p id="confirm-why" className="text-sm text-muted" aria-live="polite">
        {failing ? t("orders.checklist.missing", { count: failing }) : t("orders.checklist.ready")}
      </p>
      <ul>
        {checks.map((check) => (
          <ChecklistItem
            key={check.key}
            passed={check.passed}
            label={check.label}
            hint={check.hint}
            passedText={t("orders.checklist.passed")}
            failedText={t("orders.checklist.failed")}
          />
        ))}
      </ul>
      {canConfirm && (
        <button
          type="button"
          className={`${buttonClass} w-full`}
          aria-describedby="confirm-why"
          disabled={disabled || failing > 0}
          onClick={onConfirm}
        >
          {t("orders.editor.confirm")}
        </button>
      )}
      <p className="text-xs text-muted">{t("orders.checklist.serverNote")}</p>
    </aside>
  );
}
