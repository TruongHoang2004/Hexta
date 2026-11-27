"use client";

import { useState } from "react";
import { ToggleGroup } from "@hexta/ui";
import { ConfirmDialog, ConflictBanner, Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useReplyMode } from "../../hooks/use-reply-mode";
import type { ReplyMode } from "../../services/reply-mode";
import { ReplyModeFlow } from "./reply-mode-flow";
import { ReplyModeOption } from "./reply-mode-option";

const modes: ReplyMode[] = ["off", "suggest", "auto_send"];

const isMode = (value: string): value is ReplyMode => (modes as string[]).includes(value);

/**
 * S-19 Settings › Reply mode: off (the default, bot never runs), suggest or auto-send. Members read the mode; only
 * the owner switches it, after a confirmation.
 */
export function ReplyModeScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const state = useReplyMode(tenant?.id ?? "");
  const [confirming, setConfirming] = useState<ReplyMode | null>(null);
  const settings = state.settings;
  const current = settings?.reply_mode ?? null;

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
          <div className="space-y-2">
            <h3 id="reply-mode-legend" className="font-medium">
              {t("replyMode.legend")}
            </h3>
            <ToggleGroup
              type="single"
              aria-labelledby="reply-mode-legend"
              className="flex-col flex-nowrap gap-2"
              value={confirming ?? current ?? ""}
              disabled={!isOwner || state.saving}
              onValueChange={(value) => {
                // Pressing the selected item again yields "": keep the current mode.
                if (isMode(value) && value !== current) setConfirming(value);
              }}
            >
              {modes.map((mode) => (
                <ReplyModeOption key={mode} mode={mode} current={current === mode} />
              ))}
            </ToggleGroup>
          </div>
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
        {current === "off" && confirming && confirming !== "off" && <p>{t("replyMode.confirm.enables")}</p>}
      </ConfirmDialog>
    </div>
  );
}
