import type { ReactNode } from "react";

/** Label wrapping its control, so the control gets its accessible name from `title`. */
export function Field({ title, children }: { title: string; children: ReactNode }) {
  return (
    <label className="block space-y-1.5 text-sm font-medium">
      {title}
      {children}
    </label>
  );
}
