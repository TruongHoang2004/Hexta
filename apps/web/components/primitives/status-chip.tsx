import type { ReactNode } from "react";
import { Badge, type BadgeTone } from "@hexta/ui";

export type ChipTone = BadgeTone;

/** Small status label; meaning is carried by text, tone only reinforces it. */
export function StatusChip({ tone = "neutral", children }: { tone?: ChipTone; children: ReactNode }) {
  return <Badge tone={tone}>{children}</Badge>;
}
