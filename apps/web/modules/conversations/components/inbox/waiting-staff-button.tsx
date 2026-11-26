"use client";

import { Bell } from "lucide-react";
import { Button } from "@hexta/ui";
import { useT } from "@/lib/i18n/provider";

/** Desk header notification: how many conversations wait for staff; opens the waiting filter. */
export function WaitingStaffButton({
  count,
  active,
  onOpen,
}: {
  /** null while the first poll has not answered. */
  count: number | null;
  active: boolean;
  onOpen: () => void;
}) {
  const t = useT();
  const label = count === null ? t("desk.waiting.unknown") : t("desk.waiting.count", { count });
  return (
    <Button
      type="button"
      variant="outline"
      className={`gap-2 rounded-full ${count ? "border-warning" : ""}`}
      aria-pressed={active}
      aria-label={count === null ? label : t("desk.waiting.aria", { count })}
      onClick={onOpen}
    >
      <Bell aria-hidden="true" className="size-4" />
      <span aria-live="polite">{label}</span>
    </Button>
  );
}
