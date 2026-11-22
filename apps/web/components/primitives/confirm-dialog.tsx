import { useEffect, useRef, type ReactNode } from "react";
import { buttonClass, secondaryClass } from "./classes";

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
          <button type="button" className={buttonClass} onClick={onConfirm} disabled={busy || confirmDisabled}>
            {confirmLabel}
          </button>
        </div>
      </div>
    </dialog>
  );
}
