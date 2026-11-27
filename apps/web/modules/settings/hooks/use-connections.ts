"use client";

import { useCallback } from "react";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { connections, facebook, telegram } from "../services/channels";
import type { ChannelConnection, ChannelPlatform, ConfirmedAction } from "../types";
import { connectionName, telegramTokenPattern } from "../utils";
import { useConnectionList } from "./use-connection-list";

const confirmKey = {
  facebook: {
    pause: "channels.facebook.confirmPause",
    resume: "channels.facebook.confirmResume",
    disconnect: "channels.facebook.confirmDisconnect",
  },
  telegram: {
    pause: "channels.telegram.confirmPause",
    resume: "channels.telegram.confirmResume",
    disconnect: "channels.telegram.confirmDisconnect",
  },
} as const satisfies Record<ChannelPlatform, Record<ConfirmedAction, string>>;

/**
 * Every connection of the workspace (any number per type) with the owner actions: connect a new
 * Facebook Page or Telegram bot, pause/resume/disconnect any connection, rotate a credential where
 * the type supports it. Raw tokens are cleared by the calling forms and never kept here.
 */
export function useConnections(tenantID: string) {
  const t = useT();
  const errorText = useErrorText();
  const load = useCallback((tenant: string, signal: AbortSignal) => connections.list(tenant, signal), []);
  const list = useConnectionList<ChannelConnection>(tenantID, load);

  const statusText = (row: ChannelConnection) =>
    row.platform === "telegram" && row.status === "auth_expired"
      ? t("channels.telegram.tokenRejected")
      : t(`channels.status.${row.status}`);

  async function connectFacebook() {
    if (!tenantID || list.busy) return;
    list.setBusy("connect");
    list.setError("");
    try {
      // Leaves the app for Meta; the callback page returns here after Page selection.
      window.location.assign(await facebook.start(tenantID));
    } catch (err) {
      list.setError(errorText(err));
      list.setBusy("");
    }
  }

  function connectTelegram(token: string) {
    if (!tenantID || list.busy || !telegramTokenPattern.test(token)) return Promise.resolve(false);
    return list.mutate("connect", async () => {
      const created = await telegram.connect(tenantID, token);
      return t("channels.telegram.connected", { name: connectionName(created) });
    });
  }

  function rotate(row: ChannelConnection, token: string) {
    if (!tenantID || list.busy || !telegramTokenPattern.test(token)) return Promise.resolve(false);
    return list.mutate(row.id, async () => {
      const updated = await connections.update(tenantID, row.id, "rotate_token", token);
      return t("channels.telegram.rotated", { name: connectionName(updated) });
    });
  }

  function act(row: ChannelConnection, action: ConfirmedAction) {
    if (!tenantID || list.busy || !window.confirm(t(confirmKey[row.platform][action]))) return;
    return list.mutate(row.id, async () => {
      const updated = await connections.update(tenantID, row.id, action);
      return t("channels.nowStatus", { name: connectionName(updated), status: statusText(updated) });
    });
  }

  return { ...list, statusText, connectFacebook, connectTelegram, rotate, act };
}
