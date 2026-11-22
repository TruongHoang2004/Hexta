"use client";

import { StatusChip, buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ConnectionAction, ConnectionStatus, TelegramConnection } from "../../types";
import { connectionTone } from "../../utils";
import { ConnectionActions } from "./connection-actions";
import { TelegramRotateForm } from "./telegram-rotate-form";

interface Props {
  row: TelegramConnection;
  owner: boolean;
  busy: string;
  rotating: boolean;
  statusText: (status: ConnectionStatus) => string;
  onToggleRotate: () => void;
  onRotate: (token: string) => Promise<unknown>;
  onAct: (action: ConnectionAction) => void;
}

export function TelegramConnectionRow({ row, owner, busy, rotating, statusText, onToggleRotate, onRotate, onAct }: Props) {
  const t = useT();
  return (
    <li className="space-y-2 py-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="font-medium">{row.account_name || t("channels.telegram.fallbackName")}</p>
          <p className="break-all font-mono text-xs text-muted">
            {row.account_username ? `@${row.account_username} · ` : ""}
            {t("channels.telegram.id", { id: row.account_id })}
          </p>
          {row.last_error_message && <p className="mt-1 text-xs text-error">{row.last_error_message}</p>}
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <StatusChip tone={connectionTone[row.status]}>{statusText(row.status)}</StatusChip>
          {owner && (
            <ConnectionActions status={row.status} disabled={!!busy} saving={busy === row.id} onAct={onAct}>
              <button
                className={row.status === "auth_expired" ? buttonClass : secondaryClass}
                disabled={!!busy}
                aria-expanded={rotating}
                onClick={onToggleRotate}
              >
                {t("channels.telegram.rotate")}
              </button>
            </ConnectionActions>
          )}
        </div>
      </div>
      {owner && rotating && <TelegramRotateForm row={row} busy={busy} onRotate={onRotate} onCancel={onToggleRotate} />}
    </li>
  );
}
