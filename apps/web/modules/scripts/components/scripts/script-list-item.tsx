"use client";

import { useT } from "@/lib/i18n/provider";
import type { ScriptSummary } from "../../types";
import { ScriptStatusChip } from "./script-status-chip";

interface Props {
  script: ScriptSummary;
  selected: boolean;
  onSelect: () => void;
}

/** One list row: name, active version and status. The whole row selects the script. */
export function ScriptListItem({ script, selected, onSelect }: Props) {
  const t = useT();
  return (
    <li>
      <button
        type="button"
        aria-current={selected ? "true" : undefined}
        onClick={onSelect}
        className={`grid w-full grid-cols-[1fr_auto_auto] items-center gap-3 rounded-lg border p-3 text-left text-sm hover:bg-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 ${selected ? "border-foreground" : "border-border"}`}
      >
        <span className="min-w-0 break-words font-medium">{script.name}</span>
        <span className="text-muted">
          {script.active_version_number ? t("scripts.version", { number: script.active_version_number }) : t("scripts.none")}
        </span>
        <ScriptStatusChip status={script.status} />
      </button>
    </li>
  );
}
