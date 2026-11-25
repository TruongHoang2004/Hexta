"use client";

import { ToggleGroupItem } from "@hexta/ui";
import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ReplyMode } from "../../services/reply-mode";

/** One selectable reply mode with its plain-language explanation; an item of the reply-mode ToggleGroup. */
export function ReplyModeOption({ mode, current }: { mode: ReplyMode; current: boolean }) {
  const t = useT();
  const id = `reply-mode-${mode}`;
  return (
    <ToggleGroupItem
      value={mode}
      aria-describedby={`${id}-body`}
      className="flex h-auto w-full flex-col items-start gap-1 rounded-xl bg-background p-4 text-left text-foreground data-[state=on]:border-foreground data-[state=on]:bg-surface data-[state=on]:text-foreground data-[state=on]:ring-1 data-[state=on]:ring-foreground disabled:cursor-not-allowed disabled:opacity-70"
    >
      <span className="flex flex-wrap items-center gap-2 font-semibold">
        {t(`replyMode.modes.${mode}.name`)}
        {current && <StatusChip tone="solid">{t("replyMode.current")}</StatusChip>}
      </span>
      <span id={`${id}-body`} className="block text-sm font-normal">
        {t(`replyMode.modes.${mode}.body`)}
      </span>
      <span className="block text-xs font-normal text-muted">{t(`replyMode.modes.${mode}.fit`)}</span>
    </ToggleGroupItem>
  );
}
