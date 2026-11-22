import type { ReactNode } from "react";

export type ChipTone = "neutral" | "solid" | "warning" | "success" | "danger" | "target";
const chipTones: Record<ChipTone, string> = {
  neutral: "border-border bg-surface text-foreground",
  solid: "border-foreground bg-foreground text-background",
  warning: "border-warning bg-warning/15 text-foreground",
  success: "border-success bg-success/15 text-foreground",
  danger: "border-error bg-error/10 text-foreground",
  target: "border-warning border-dashed bg-surface text-muted",
};

/** Small status label; meaning is carried by text, tone only reinforces it. */
export function StatusChip({ tone = "neutral", children }: { tone?: ChipTone; children: ReactNode }) {
  return (
    <span
      className={`inline-flex items-center whitespace-nowrap rounded-full border px-2 py-0.5 text-xs font-medium ${chipTones[tone]}`}
    >
      {children}
    </span>
  );
}
