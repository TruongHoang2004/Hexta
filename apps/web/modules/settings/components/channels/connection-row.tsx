"use client";

import { StatusChip, buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ChannelConnection, ConfirmedAction } from "../../types";
import { connectionTone, platformLabel } from "../../utils";
import { ConnectionActions } from "./connection-actions";
import { TelegramRotateForm } from "./telegram-rotate-form";

interface Props {
  row: ChannelConnection;
  owner: boolean;
  busy: string;
  statusText: string;
  rotating: boolean;
  onToggleRotate: () => void;
  onRotate: (token: string) => Promise<unknown>;
  onReconnect: () => void;
  onAct: (action: ConfirmedAction) => void;
}

/** One connection of any type: type, account, status and the actions the API allows. */
export function ConnectionRow({ row, owner, busy, statusText, rotating, onToggleRotate, onRotate, onReconnect, onAct }: Props) {
  const t = useT();
  const expired = row.status === "auth_expired";
  return (
    <li className="space-y-2 py-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="font-medium">
            <span className="mr-2 rounded-md border border-border px-1.5 py-0.5 text-xs text-muted">
              {platformLabel[row.platform]}
            </span>
            {row.account_name || t(`channels.${row.platform}.fallbackName`)}
          </p>
          <p className="mt-1 break-all font-mono text-xs text-muted">
            {row.account_username ? `@${row.account_username} · ` : ""}
            {t("channels.accountID", { id: row.account_id })}
          </p>
          {row.last_error_message && <p className="mt-1 text-xs text-error">{row.last_error_message}</p>}
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <StatusChip tone={connectionTone[row.status]}>{statusText}</StatusChip>
          {owner && (
            <ConnectionActions actions={row.actions} disabled={!!busy} saving={busy === row.id} onAct={onAct}>
              {row.actions.includes("rotate_token") && (
                <button
                  className={expired ? buttonClass : secondaryClass}
                  disabled={!!busy}
                  aria-expanded={rotating}
                  onClick={onToggleRotate}
                >
                  {t("channels.telegram.rotate")}
                </button>
              )}
              {row.platform === "facebook" && expired && (
                <button className={buttonClass} disabled={!!busy} onClick={onReconnect}>
                  {t("channels.facebook.reconnect")}
                </button>
              )}
            </ConnectionActions>
          )}
        </div>
      </div>
      {owner && rotating && <TelegramRotateForm row={row} busy={busy} onRotate={onRotate} onCancel={onToggleRotate} />}
    </li>
  );
}
