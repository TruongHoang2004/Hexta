import { useEffect, useRef, type ReactNode } from "react";
export const inputClass =
  "w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60";
export const buttonClass =
  "rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50";
export const secondaryClass =
  "rounded-lg border border-border px-3 py-2 text-sm hover:bg-surface disabled:opacity-50";
export function Field({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <label className="block space-y-1.5 text-sm font-medium">
      {title}
      {children}
    </label>
  );
}
export function Notice({
  error,
  success,
}: {
  error?: string;
  success?: string;
}) {
  return (
    <>
      {error && (
        <p
          role="alert"
          className="rounded-lg border border-red-300 bg-red-50 p-3 text-sm text-red-800"
        >
          {error}
        </p>
      )}
      {success && (
        <p
          role="status"
          className="rounded-lg border border-green-300 bg-green-50 p-3 text-sm text-green-800"
        >
          {success}
        </p>
      )}
    </>
  );
}
export function Pager({
  page,
  total,
  onChange,
  disabled,
}: {
  page: number;
  total: number;
  onChange: (page: number) => void;
  disabled?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-3 border-t border-border pt-3 text-sm">
      <span>
        {total} records · Page {page} of {Math.max(1, Math.ceil(total / 20))}
      </span>
      <div className="flex gap-2">
        <button
          type="button"
          className={secondaryClass}
          disabled={disabled || page <= 1}
          onClick={() => onChange(page - 1)}
        >
          Previous
        </button>
        <button
          type="button"
          className={secondaryClass}
          disabled={disabled || page * 20 >= total}
          onClick={() => onChange(page + 1)}
        >
          Next
        </button>
      </div>
    </div>
  );
}

export type ChipTone = "neutral" | "solid" | "warning" | "success" | "danger" | "target";
const chipTones: Record<ChipTone, string> = {
  neutral: "border-border bg-surface text-foreground",
  solid: "border-foreground bg-foreground text-background",
  warning: "border-warning bg-warning/15 text-foreground",
  success: "border-success bg-success/15 text-foreground",
  danger: "border-error bg-error/10 text-foreground",
  target: "border-warning border-dashed bg-surface text-muted",
};
/** Small status label; meaning is carried by text, tone only reinforces it. */
export function StatusChip({ tone = "neutral", children }: { tone?: ChipTone; children: ReactNode }) {
  return (
    <span
      className={`inline-flex items-center whitespace-nowrap rounded-full border px-2 py-0.5 text-xs font-medium ${chipTones[tone]}`}
    >
      {children}
    </span>
  );
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-xl border border-dashed border-border p-8 text-center">
      <p className="font-semibold">{title}</p>
      {description && <p className="max-w-md text-sm text-muted">{description}</p>}
      {action && <div className="mt-2">{action}</div>}
    </div>
  );
}

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

export interface FilterOption<T extends string> {
  value: T;
  label: string;
  count?: number;
}
/** Single-choice filter rendered as toggle buttons. */
export function SegmentedFilter<T extends string>({
  label,
  options,
  value,
  onChange,
}: {
  label: string;
  options: FilterOption<T>[];
  value: T;
  onChange: (value: T) => void;
}) {
  return (
    <div role="group" aria-label={label} className="flex flex-wrap gap-1.5">
      {options.map((option) => {
        const active = option.value === value;
        return (
          <button
            key={option.value}
            type="button"
            aria-pressed={active}
            onClick={() => onChange(option.value)}
            className={`min-h-9 rounded-full border px-3 text-sm ${active ? "border-foreground bg-foreground text-background" : "border-border bg-surface hover:bg-background"}`}
          >
            {option.label}
            {option.count !== undefined && <span className="ml-1.5 opacity-80">{option.count}</span>}
          </button>
        );
      })}
    </div>
  );
}

export interface TabItem<T extends string> {
  value: T;
  label: string;
  badge?: ReactNode;
}
/** ARIA tablist with arrow-key navigation; the caller renders the active panel. */
export function Tabs<T extends string>({
  label,
  items,
  value,
  onChange,
  idPrefix,
}: {
  label: string;
  items: TabItem<T>[];
  value: T;
  onChange: (value: T) => void;
  idPrefix: string;
}) {
  function onKeyDown(event: React.KeyboardEvent<HTMLButtonElement>, index: number) {
    const delta = event.key === "ArrowRight" ? 1 : event.key === "ArrowLeft" ? -1 : 0;
    if (!delta) return;
    event.preventDefault();
    const next = items[(index + delta + items.length) % items.length];
    onChange(next.value);
    document.getElementById(`${idPrefix}-tab-${next.value}`)?.focus();
  }
  return (
    <div role="tablist" aria-label={label} className="flex flex-wrap gap-1 border-b border-border">
      {items.map((item, index) => {
        const active = item.value === value;
        return (
          <button
            key={item.value}
            id={`${idPrefix}-tab-${item.value}`}
            type="button"
            role="tab"
            aria-selected={active}
            aria-controls={`${idPrefix}-panel`}
            tabIndex={active ? 0 : -1}
            onClick={() => onChange(item.value)}
            onKeyDown={(event) => onKeyDown(event, index)}
            className={`-mb-px inline-flex min-h-11 items-center gap-2 border-b-2 px-3 text-sm ${active ? "border-foreground font-semibold" : "border-transparent text-muted hover:text-foreground"}`}
          >
            {item.label}
            {item.badge}
          </button>
        );
      })}
    </div>
  );
}

/** Modal built on the native dialog element (focus trap and Escape handled by the browser). */
export function ConfirmDialog({
  open,
  title,
  children,
  confirmLabel,
  cancelLabel,
  onConfirm,
  onCancel,
  busy,
  confirmDisabled,
}: {
  open: boolean;
  title: string;
  children: ReactNode;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
  busy?: boolean;
  confirmDisabled?: boolean;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  }, [open]);
  return (
    <dialog
      ref={ref}
      aria-labelledby="confirm-dialog-title"
      onCancel={(event) => {
        event.preventDefault();
        onCancel();
      }}
      className="m-auto w-[min(32rem,calc(100vw-2rem))] rounded-xl border border-border bg-surface p-0 text-foreground backdrop:bg-black/40"
    >
      <div className="space-y-4 p-5">
        <h2 id="confirm-dialog-title" className="text-lg font-semibold">
          {title}
        </h2>
        <div className="space-y-3 text-sm">{children}</div>
        <div className="flex flex-wrap justify-end gap-2">
          <button type="button" className={secondaryClass} onClick={onCancel} disabled={busy}>
            {cancelLabel}
          </button>
          <button
            type="button"
            className={buttonClass}
            onClick={onConfirm}
            disabled={busy || confirmDisabled}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </dialog>
  );
}

/** One row of a pass/fail checklist (e.g. order confirmation conditions). */
export function ChecklistItem({
  passed,
  label,
  hint,
  passedText,
  failedText,
}: {
  passed: boolean;
  label: string;
  hint?: string;
  passedText: string;
  failedText: string;
}) {
  return (
    <li className="flex items-start gap-2.5 border-b border-border py-2 text-sm last:border-b-0">
      <span
        aria-hidden="true"
        className={`mt-0.5 inline-flex h-5 w-5 flex-none items-center justify-center rounded-full text-xs font-bold ${passed ? "bg-success text-white" : "border-2 border-error text-error"}`}
      >
        {passed ? "✓" : "!"}
      </span>
      <span>
        <span className="sr-only">{passed ? passedText : failedText}: </span>
        <span className={passed ? "" : "font-semibold"}>{label}</span>
        {hint && <span className="block text-muted">{hint}</span>}
      </span>
    </li>
  );
}
