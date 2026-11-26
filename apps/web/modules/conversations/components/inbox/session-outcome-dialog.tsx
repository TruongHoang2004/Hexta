"use client";

import { useState } from "react";
import { Textarea } from "@hexta/ui";
import { ConfirmDialog, Picker } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { SessionInput } from "../../services/messaging";
import { maxSessionNote, sessionOutcomes, type SessionOutcome } from "../../types";
import { isOneOf } from "../../utils";

/**
 * Asks for a session outcome (required) and an optional bounded note: closing the current session (`mode="close"`)
 * or editing a closed session's outcome (`mode="outcome"`). Mount it with a key per session so fields reset.
 */
export function SessionOutcomeDialog({
  open,
  mode,
  initial,
  busy,
  error,
  onCancel,
  onConfirm,
}: {
  open: boolean;
  mode: "close" | "outcome";
  initial?: Partial<SessionInput>;
  busy: boolean;
  /** Failure of the last confirm, shown inside the dialog. */
  error?: string;
  onCancel: () => void;
  onConfirm: (input: SessionInput) => void;
}) {
  const t = useT();
  const [outcome, setOutcome] = useState<SessionOutcome | "">(initial?.outcome ?? "");
  const [note, setNote] = useState(initial?.note ?? "");
  const copy = mode === "close" ? "closeDialog" : "outcomeDialog";
  const tooLong = note.trim().length > maxSessionNote;
  return (
    <ConfirmDialog
      open={open}
      title={t(`desk.sessions.${copy}.title`)}
      confirmLabel={t(`desk.sessions.${copy}.confirm`)}
      cancelLabel={t("common.cancel")}
      busy={busy}
      confirmDisabled={!outcome || tooLong}
      onCancel={onCancel}
      onConfirm={() => {
        if (outcome) onConfirm({ outcome, note: note.trim() });
      }}
    >
      <p>{t(`desk.sessions.${copy}.body`)}</p>
      {error && (
        <p role="alert" className="text-sm text-error">
          {error}
        </p>
      )}
      <label className="block space-y-1.5 text-sm">
        <span className="font-medium">{t("desk.sessions.outcomeLabel")}</span>
        <Picker
          required
          disabled={busy}
          value={outcome}
          onChange={(value) => setOutcome(isOneOf(sessionOutcomes, value) ? value : "")}
          options={sessionOutcomes.map((code) => ({ value: code, label: t(`desk.sessions.outcomes.${code}`) }))}
        />
      </label>
      <label className="block space-y-1.5 text-sm">
        <span className="font-medium">{t("desk.sessions.noteLabel")}</span>
        <Textarea
          value={note}
          disabled={busy}
          maxLength={maxSessionNote}
          aria-describedby="session-note-count"
          onChange={(event) => setNote(event.target.value)}
        />
        <span id="session-note-count" className={`block text-xs ${tooLong ? "text-error" : "text-muted"}`}>
          {t("desk.sessions.noteCount", { count: note.trim().length, max: maxSessionNote })}
        </span>
      </label>
    </ConfirmDialog>
  );
}
