import { Suspense } from "react";
import { FacebookCallback } from "@/components/channels/facebook-callback";

export default function FacebookCallbackPage() {
  return (
    <Suspense fallback={<p role="status">Loading…</p>}>
      <FacebookCallback />
    </Suspense>
  );
}
