"use client";

import { useEffect, useRef } from "react";
import { useSearchParams } from "next/navigation";

/**
 * Opens the work item named by `?item=<id>` (links from the Overview) once per ID. Renders nothing;
 * wrap it in a Suspense boundary because it reads search params.
 */
export function WorkDeepLink({ onOpen }: { onOpen: (id: string) => void }) {
  const id = useSearchParams().get("item");
  const opened = useRef<string | null>(null);
  useEffect(() => {
    if (id && opened.current !== id) {
      opened.current = id;
      onOpen(id);
    }
  }, [id, onOpen]);
  return null;
}
