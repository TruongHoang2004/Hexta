"use client";

import { StatusChip, type ChipTone } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { KnowledgeSyncState } from "../types";

const tones: Record<Exclude<KnowledgeSyncState, "">, ChipTone> = { pending: "warning", done: "success", failed: "danger" };

/** Whether the bot can already answer from a document (the AI index state); text carries the meaning. */
export function KnowledgeSyncChip({ state, error }: { state: KnowledgeSyncState; error?: string }) {
  const t = useT();
  const key = state || "pending";
  return (
    <span className="flex flex-wrap items-center gap-1">
      <StatusChip tone={tones[key]}>{t(`knowledge.sync.${key}`)}</StatusChip>
      {key === "failed" && error && <span className="text-xs text-muted">{t("knowledge.sync.reason", { code: error })}</span>}
    </span>
  );
}
