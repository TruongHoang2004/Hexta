"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { WorkItem } from "../../types";
import { stateTone } from "../../utils";

interface Props {
  item: WorkItem;
  person: (userID: string | null | undefined) => string;
  selected: boolean;
  busy: boolean;
  onOpen: () => void;
}

export function WorkListItem({ item, person, selected, busy, onOpen }: Props) {
  const t = useT();
  const step = item.workflow.steps[item.current_step];
  return (
    <button
      type="button"
      disabled={busy}
      aria-current={selected ? "true" : undefined}
      className={`block w-full space-y-2 rounded-lg border p-3 text-left hover:bg-background ${selected ? "border-foreground" : "border-border"}`}
      onClick={onOpen}
    >
      <span className="flex flex-wrap items-center justify-between gap-2">
        <span className="font-medium">{item.workflow.name}</span>
        <StatusChip tone={stateTone[item.state]}>{t(`work.states.${item.state}`)}</StatusChip>
      </span>
      <span className="block text-sm text-muted">
        {t("work.step", { current: item.current_step + 1, total: item.workflow.steps.length, name: step?.name ?? "" })}
      </span>
      <span className="block break-all text-xs">
        {t("work.responsible")}: {person(item.responsible_id)}
        {item.recipient_id && item.state === "waiting"
          ? ` · ${t("work.waitingShort", { name: person(item.recipient_id) })}`
          : ""}
      </span>
    </button>
  );
}
