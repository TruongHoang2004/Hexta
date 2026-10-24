import { Suspense } from "react";
import { OwnerOverview } from "@/components/sales/overview";

export default function Page() {
  return (
    <Suspense fallback={<p role="status">Loading overview…</p>}>
      <OwnerOverview />
    </Suspense>
  );
}
