"use client";

import { useState } from "react";
import { ConfirmDialog, ConflictBanner, Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useReplyMode } from "../../hooks/use-reply-mode";
import type { ReplyMode } from "../../services/reply-mode";
import { ReplyModeFlow } from "./reply-mode-flow";
import { ReplyModeOption } from "./reply-mode-option";

const modes: ReplyMode[] = ["suggest", "auto_send"];

/** S-19 Settings › Reply mode. Members read the mode; only the owner switches it, after a confirmation. */
export function ReplyModeScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const state = useReplyMode(tenant?.id ?? "");
  const [confirming, setConfirming] = useState<ReplyMode | null>(null);
  const settings = state.settings;
  const current = settings?.configured ? settings.reply_mode : null;

  return (
    <div className="max-w-3xl space-y-4">
      <div>
        <h2 className="text-lg font-semibold">{t("replyMode.title")}</h2>
        <p className="mt-1 text-sm text-muted">{t("replyMode.intro")}</p>
      </div>
      {state.conflict ? (
        <ConflictBanner message={t("replyMode.conflict")} actionLabel={t("replyMode.reload")} onReload={state.reload} />
      ) : (
        <Notice error={state.error} success={state.saved ? t("replyMode.saved", { mode: t(`replyMode.modes.${state.saved}.name`) }) : ""} />
      )}
      {!isOwner && <p className="text-sm text-muted">{t("replyMode.ownerOnly")}</p>}
      {state.loading && !settings ? (
        <p role="status">{t("replyMode.loading")}</p>
      ) : (
        settings && (
          <>
            {!settings.configured && (
              <p role="status" className="rounded-lg border border-warning bg-warning/15 p-3 text-sm">
                {t("replyMode.notConfigured")}
              </p>
            )}
            <fieldset className="space-y-2">
              <legend className="mb-2 font-medium">{t("replyMode.legend")}</legend>
              {modes.map((mode) => (
                <ReplyModeOption
                  key={mode}
                  mode={mode}
                  current={current === mode}
                  checked={(confirming ?? current) === mode}
                  disabled={!isOwner || state.saving}
                  onSelect={() => {
                    if (mode !== current) setConfirming(mode);
                  }}
                />
              ))}
            </fieldset>
          </>
        )
      )}
      <ReplyModeFlow />
      <ConfirmDialog
        open={confirming !== null}
        title={confirming ? t(`replyMode.confirm.${confirming}.title`) : ""}
        confirmLabel={confirming ? t(`replyMode.confirm.${confirming}.action`) : ""}
        cancelLabel={t("common.cancel")}
        busy={state.saving}
        onCancel={() => setConfirming(null)}
        onConfirm={() => {
          if (!confirming) return;
          void state.save(confirming).then(() => setConfirming(null));
        }}
      >
        <p>{confirming ? t(`replyMode.confirm.${confirming}.body`) : ""}</p>
        {!settings?.configured && <p>{t("replyMode.confirm.enables")}</p>}
      </ConfirmDialog>
    </div>
  );
}
