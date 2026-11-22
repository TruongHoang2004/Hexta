"use client";

import { StatusChip, buttonClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ConnectionAction, FacebookConnection } from "../../types";
import { connectionTone } from "../../utils";
import { ConnectionActions } from "./connection-actions";

interface Props {
  row: FacebookConnection;
  owner: boolean;
  busy: string;
  onReconnect: () => void;
  onAct: (action: ConnectionAction) => void;
}

export function FacebookConnectionRow({ row, owner, busy, onReconnect, onAct }: Props) {
  const t = useT();
  return (
    <li className="flex flex-wrap items-center justify-between gap-3 py-3">
      <div className="min-w-0">
        <p className="font-medium">{row.account_name || t("channels.facebook.fallbackName")}</p>
        <p className="break-all font-mono text-xs text-muted">{row.account_id}</p>
        {row.last_error_message && <p className="mt-1 text-xs text-error">{row.last_error_message}</p>}
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <StatusChip tone={connectionTone[row.status]}>{t(`channels.status.${row.status}`)}</StatusChip>
        {owner && (
          <ConnectionActions status={row.status} disabled={!!busy} saving={busy === row.id} onAct={onAct}>
            {row.status === "auth_expired" && (
              <button className={buttonClass} disabled={!!busy} onClick={onReconnect}>
                {t("channels.facebook.reconnect")}
              </button>
            )}
          </ConnectionActions>
        )}
      </div>
    </li>
  );
}
