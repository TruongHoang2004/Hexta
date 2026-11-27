"use client";

import { useState } from "react";
import { Button, Textarea } from "@hexta/ui";
import { useT } from "@/lib/i18n/provider";

/** A pending proposal's text with send, edit-then-send and discard. Editing never sends until confirmed. */
export function ProposalEditor({
  proposalID,
  text,
  canSend,
  canDiscard,
  sendHint,
  busy,
  uncertain,
  onSend,
  onDiscard,
}: {
  proposalID: string;
  text: string;
  canSend: boolean;
  canDiscard: boolean;
  sendHint: string;
  busy: boolean;
  uncertain: boolean;
  onSend: (text?: string) => void;
  onDiscard: () => void;
}) {
  const t = useT();
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(text);
  const fieldID = `proposal-text-${proposalID}`;
  const hintID = `proposal-hint-${proposalID}`;
  const edited = editing && draft.trim() !== text;
  return (
    <div className="space-y-2">
      <label htmlFor={fieldID} className="block text-xs font-medium text-muted">
        {editing ? t("desk.proposal.editLabel") : t("desk.proposal.textLabel")}
      </label>
      <Textarea
        id={fieldID}
        rows={3}
        maxLength={2000}
        value={editing ? draft : text}
        readOnly={!editing}
        disabled={busy || uncertain}
        onChange={(e) => setDraft(e.target.value)}
        aria-describedby={sendHint ? hintID : undefined}
      />
      {sendHint && (
        <p id={hintID} className="text-xs text-muted">
          {sendHint}
        </p>
      )}
      <div className="flex flex-wrap justify-end gap-2">
        {editing ? (
          <>
            <Button
              type="button"
              variant="outline"
              disabled={busy || uncertain}
              onClick={() => {
                setEditing(false);
                setDraft(text);
              }}
            >
              {t("desk.proposal.cancelEdit")}
            </Button>
            <Button type="button" disabled={busy || !canSend || !draft.trim()} onClick={() => onSend(draft)}>
              {uncertain ? t("desk.proposal.retry") : edited ? t("desk.proposal.sendEdited") : t("desk.proposal.send")}
            </Button>
          </>
        ) : (
          <>
            <Button type="button" variant="outline" disabled={busy || uncertain || !canDiscard} onClick={onDiscard}>
              {t("desk.proposal.discard")}
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={busy || uncertain || !canSend}
              onClick={() => {
                setDraft(text);
                setEditing(true);
              }}
            >
              {t("desk.proposal.edit")}
            </Button>
            <Button type="button" disabled={busy || !canSend} onClick={() => onSend()}>
              {uncertain ? t("desk.proposal.retry") : t("desk.proposal.send")}
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
