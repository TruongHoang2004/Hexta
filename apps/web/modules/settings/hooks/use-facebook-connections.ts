"use client";

import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { channels } from "../services/channels";
import type { ConnectionAction, FacebookConnection } from "../types";
import { useConnectionList } from "./use-connection-list";

const confirmKey = {
  pause: "channels.facebook.confirmPause",
  resume: "channels.facebook.confirmResume",
  disconnect: "channels.facebook.confirmDisconnect",
} as const satisfies Record<ConnectionAction, string>;

const loadPages = (tenant: string, signal: AbortSignal) => channels.list(tenant, signal);

/** Connected Facebook Pages plus OAuth start and pause/resume/disconnect. */
export function useFacebookConnections(tenantID: string) {
  const t = useT();
  const errorText = useErrorText();
  const list = useConnectionList(tenantID, loadPages);

  async function connect() {
    if (!tenantID || list.busy) return;
    list.setBusy("connect");
    list.setError("");
    try {
      // Leaves the app for Meta; the callback page returns here after Page selection.
      window.location.assign(await channels.startFacebook(tenantID));
    } catch (err) {
      list.setError(errorText(err));
      list.setBusy("");
    }
  }

  function act(row: FacebookConnection, action: ConnectionAction) {
    if (!tenantID || list.busy || !window.confirm(t(confirmKey[action]))) return;
    return list.mutate(row.id, async () => {
      const updated = await channels.update(tenantID, row.id, action);
      return t("channels.nowStatus", {
        name: updated.account_name || updated.account_id,
        status: t(`channels.status.${updated.status}`),
      });
    });
  }

  return { ...list, connect, act };
}
