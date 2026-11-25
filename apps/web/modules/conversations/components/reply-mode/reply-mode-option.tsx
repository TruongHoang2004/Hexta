"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ReplyMode } from "../../services/reply-mode";

/** One selectable reply mode with its plain-language explanation. */
export function ReplyModeOption({
  mode,
  current,
  checked,
  disabled,
  onSelect,
}: {
  mode: ReplyMode;
  current: boolean;
  checked: boolean;
  disabled: boolean;
  onSelect: () => void;
}) {
  const t = useT();
  const id = `reply-mode-${mode}`;
  return (
    <label
      htmlFor={id}
      className={`flex cursor-pointer gap-3 rounded-xl border p-4 has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-primary/40 ${checked ? "border-foreground bg-surface" : "border-border bg-background"} ${disabled ? "cursor-not-allowed opacity-70" : ""}`}
    >
      <input
        id={id}
        type="radio"
        name="reply-mode"
        value={mode}
        className="mt-1"
        checked={checked}
        disabled={disabled}
        onChange={onSelect}
        aria-describedby={`${id}-body`}
      />
      <span className="space-y-1">
        <span className="flex flex-wrap items-center gap-2 font-semibold">
          {t(`replyMode.modes.${mode}.name`)}
          {current && <StatusChip tone="solid">{t("replyMode.current")}</StatusChip>}
        </span>
        <span id={`${id}-body`} className="block text-sm">
          {t(`replyMode.modes.${mode}.body`)}
        </span>
        <span className="block text-xs text-muted">{t(`replyMode.modes.${mode}.fit`)}</span>
      </span>
    </label>
  );
}
