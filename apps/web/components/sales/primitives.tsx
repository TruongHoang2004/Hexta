import type { ReactNode } from "react";
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
