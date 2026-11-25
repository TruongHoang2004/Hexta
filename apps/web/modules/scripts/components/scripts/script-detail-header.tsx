"use client";

import { useState } from "react";
import { ConfirmDialog, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { Script } from "../../types";
import { activeVersion, draftNumber } from "../../utils";
import { ScriptStatusChip } from "./script-status-chip";

interface Props {
  script: Script;
  canActivate: boolean;
  canArchive: boolean;
  busy: boolean;
  onDeactivate: () => void;
  onArchive: () => Promise<boolean>;
}

/** Script name, status line and the deactivate/archive actions the caller is allowed to use. */
export function ScriptDetailHeader({ script, canActivate, canArchive, busy, onDeactivate, onArchive }: Props) {
  const t = useT();
  const [confirming, setConfirming] = useState(false);
  const active = activeVersion(script);
  const note =
    script.status === "active" && active
      ? t("scripts.detail.activeNote", { active: active.number, draft: draftNumber(script) })
      : script.status === "archived"
        ? t("scripts.detail.archivedNote")
        : t("scripts.detail.inactiveNote");
  const archived = script.status === "archived";
  return (
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div className="min-w-0 space-y-1">
        <h2 className="flex flex-wrap items-center gap-2 text-lg font-semibold">
          <span className="break-words">{script.name}</span>
          <ScriptStatusChip status={script.status} />
        </h2>
        <p className="text-sm text-muted">{note}</p>
      </div>
      {!archived && (
        <div className="flex flex-wrap gap-2">
          {canActivate && script.status === "active" && (
            <button type="button" className={secondaryClass} disabled={busy} onClick={onDeactivate}>
              {t("scripts.detail.deactivate")}
            </button>
          )}
          {canArchive && (
            <button type="button" className={secondaryClass} disabled={busy} onClick={() => setConfirming(true)}>
              {t("scripts.detail.archive")}
            </button>
          )}
        </div>
      )}
      <ConfirmDialog
        open={confirming}
        title={t("scripts.detail.archiveTitle", { name: script.name })}
        confirmLabel={t("scripts.detail.archive")}
        cancelLabel={t("common.cancel")}
        busy={busy}
        onCancel={() => setConfirming(false)}
        onConfirm={() => {
          void onArchive().then(() => setConfirming(false));
        }}
      >
        <p>{t("scripts.detail.archiveBody")}</p>
      </ConfirmDialog>
    </div>
  );
}
