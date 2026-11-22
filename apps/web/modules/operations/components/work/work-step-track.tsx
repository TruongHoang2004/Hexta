"use client";

import { useT } from "@/lib/i18n/provider";
import type { WorkItem } from "../../types";

/** Compact step progress: done steps, current step highlighted, later steps muted. */
export function WorkStepTrack({ item }: { item: WorkItem }) {
  const t = useT();
  return (
    <ol aria-label={t("work.progress")} className="flex flex-wrap items-center gap-1 text-xs">
      {item.workflow.steps.map((step) => {
        const done = step.ordinal < item.current_step || item.state === "completed";
        const current = step.ordinal === item.current_step && item.state !== "completed";
        return (
          <li
            key={step.id}
            aria-current={current ? "step" : undefined}
            className={`rounded px-2 py-1 ${current ? "bg-foreground font-semibold text-background" : done ? "bg-border text-muted line-through" : "bg-background text-muted"}`}
          >
            {step.ordinal + 1}. {step.name}
          </li>
        );
      })}
    </ol>
  );
}
