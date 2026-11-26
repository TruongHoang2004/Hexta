"use client";

import { buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OrderCaseType } from "../../types";

interface Props {
  busy: boolean;
  onOpen: (type: OrderCaseType) => void;
}

/** Entry point on a confirmed order without a case: open a cancellation or a return. */
export function OrderCaseOpenActions({ busy, onOpen }: Props) {
  const t = useT();
  return (
    <div className="space-y-3">
      <p className="text-sm text-muted">{t("orderCase.openPrompt")}</p>
      <div className="flex flex-wrap gap-2">
        <button type="button" className={secondaryClass} disabled={busy} onClick={() => onOpen("cancellation")}>
          {t("orderCase.openCancellation")}
        </button>
        <button type="button" className={buttonClass} disabled={busy} onClick={() => onOpen("return")}>
          {t("orderCase.openReturn")}
        </button>
      </div>
    </div>
  );
}
