"use client";

import { useRef } from "react";

/**
 * Idempotency key that stays the same while the payload is unchanged, so retrying a failed
 * submit cannot create a duplicate; any edit to the payload gets a fresh key.
 */
export function useRetryKey() {
  const retry = useRef({ payload: "", key: "" });
  return (input: unknown) => {
    const payload = JSON.stringify(input);
    if (retry.current.payload !== payload) retry.current = { payload, key: crypto.randomUUID() };
    return retry.current.key;
  };
}
