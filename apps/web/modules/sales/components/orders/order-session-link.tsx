"use client";

import { useEffect } from "react";
import { useSearchParams } from "next/navigation";

/**
 * Reports the conversation session named by `?turn=<id>` (the Desk's "Create draft order" link) so the next new
 * draft is linked to it. Renders nothing; wrap it in a Suspense boundary because it reads search params.
 */
export function OrderSessionLink({ onLink }: { onLink: (turnID: string) => void }) {
  const turn = useSearchParams().get("turn") ?? "";
  useEffect(() => {
    if (turn) onLink(turn);
  }, [turn, onLink]);
  return null;
}
