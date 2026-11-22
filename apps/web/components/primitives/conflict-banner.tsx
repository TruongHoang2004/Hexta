import { secondaryClass } from "./classes";

/** Shown when a save was rejected because someone else changed the record first. */
export function ConflictBanner({
  message,
  actionLabel,
  onReload,
}: {
  message: string;
  actionLabel: string;
  onReload: () => void;
}) {
  return (
    <div
      role="alert"
      className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-warning bg-warning/15 p-3 text-sm"
    >
      <span>{message}</span>
      <button type="button" className={secondaryClass} onClick={onReload}>
        {actionLabel}
      </button>
    </div>
  );
}
