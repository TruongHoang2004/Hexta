"use client";

import { useEffect, useRef } from "react";
import { useSearchParams } from "next/navigation";

/**
 * Opens the order named by `?order=<id>` (links from the Overview) once per ID. Renders nothing;
 * wrap it in a Suspense boundary because it reads search params.
 */
export function OrderDeepLink({ onOpen }: { onOpen: (id: string) => void }) {
  const id = useSearchParams().get("order");
  const opened = useRef<string | null>(null);
  useEffect(() => {
    if (id && opened.current !== id) {
      opened.current = id;
      onOpen(id);
    }
  }, [id, onOpen]);
  return null;
}
