"use client";

import { useT } from "@/lib/i18n/provider";
import { telegram } from "../services/channels";
import type { ConnectionAction, ConnectionStatus, TelegramConnection } from "../types";
import { botName, telegramTokenPattern } from "../utils";
import { useConnectionList } from "./use-connection-list";

const confirmKey = {
  pause: "channels.telegram.confirmPause",
  resume: "channels.telegram.confirmResume",
  disconnect: "channels.telegram.confirmDisconnect",
} as const satisfies Record<ConnectionAction, string>;

const loadBots = (tenant: string, signal: AbortSignal) => telegram.list(tenant, signal);

/**
 * Connected Telegram bots plus connect, token rotation and pause/resume/disconnect. Callers clear
 * raw tokens from their own state before calling; tokens are never kept here.
 */
export function useTelegramConnections(tenantID: string) {
  const t = useT();
  const list = useConnectionList(tenantID, loadBots);
  const statusText = (status: ConnectionStatus) =>
    status === "auth_expired" ? t("channels.telegram.tokenRejected") : t(`channels.status.${status}`);

  function connect(token: string) {
    if (!tenantID || list.busy || !telegramTokenPattern.test(token)) return Promise.resolve(false);
    return list.mutate("connect", async () => {
      const created = await telegram.connect(tenantID, token);
      return t("channels.telegram.connected", { name: botName(created) });
    });
  }

  function rotate(row: TelegramConnection, token: string) {
    if (!tenantID || list.busy || !telegramTokenPattern.test(token)) return Promise.resolve(false);
    return list.mutate(row.id, async () => {
      const updated = await telegram.update(tenantID, row.id, "rotate_token", token);
      return t("channels.telegram.rotated", { name: botName(updated) });
    });
  }

  function act(row: TelegramConnection, action: ConnectionAction) {
    if (!tenantID || list.busy || !window.confirm(t(confirmKey[action]))) return;
    return list.mutate(row.id, async () => {
      const updated = await telegram.update(tenantID, row.id, action);
      return t("channels.nowStatus", { name: botName(updated), status: statusText(updated.status) });
    });
  }

  return { ...list, statusText, connect, rotate, act };
}
