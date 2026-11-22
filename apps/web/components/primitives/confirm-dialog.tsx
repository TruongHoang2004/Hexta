"use client";

import type { ReactNode } from "react";
import { Button, Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle } from "@hexta/ui";

/** Confirmation modal (Radix Dialog: focus trap, Escape and outside click call onCancel). */
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
  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next && !busy) onCancel();
      }}
    >
      <DialogContent>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription asChild>
          <div className="space-y-3 text-sm text-foreground">{children}</div>
        </DialogDescription>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={onCancel} disabled={busy}>
            {cancelLabel}
          </Button>
          <Button type="button" onClick={onConfirm} disabled={busy || confirmDisabled}>
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
