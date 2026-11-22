import { Suspense } from "react";
import { FacebookCallbackScreen } from "@/modules/settings";

export default function FacebookCallbackPage() {
  return (
    <Suspense fallback={<p role="status" aria-busy="true">…</p>}>
      <FacebookCallbackScreen />
    </Suspense>
  );
}
