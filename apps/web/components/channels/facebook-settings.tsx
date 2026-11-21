"use client";

import { useEffect, useState } from "react";
import { useWorkspace } from "@/components/sales/workspace";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import {
  channels,
  type ConnectionAction,
  type ConnectionStatus,
  type FacebookConnection,
} from "@/lib/services/channels";
import {
  buttonClass,
  Notice,
  StatusChip,
  secondaryClass,
  type ChipTone,
} from "@/components/sales/primitives";
import { ExpiredConnectionBanner } from "@/components/channels/expired-banner";

export const statusTone: Record<ConnectionStatus, ChipTone> = {
  active: "success",
  paused: "warning",
  auth_expired: "danger",
  disconnected: "neutral",
};
const confirmKey = {
  pause: "channels.facebook.confirmPause",
  resume: "channels.facebook.confirmResume",
  disconnect: "channels.facebook.confirmDisconnect",
} as const satisfies Record<ConnectionAction, string>;

export function FacebookChannelSettings() {
  const { tenant, isOwner: owner } = useWorkspace();
  const t = useT();
  const errorText = useErrorText();
  const [rows, setRows] = useState<FacebookConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [revision, setRevision] = useState(0);
  const tenantID = tenant?.id ?? "";
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const result = await channels.list(tenantID, controller.signal);
        if (!controller.signal.aborted) setRows(result);
      } catch (err) {
        if (!controller.signal.aborted) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenantID, revision, errorText]);
  async function connect() {
    if (!tenantID || busy) return;
    setBusy("connect");
    setError("");
    try {
      // Leaves the app for Meta; the callback page returns here after Page selection.
      window.location.assign(await channels.startFacebook(tenantID));
    } catch (err) {
      setError(errorText(err));
      setBusy("");
    }
  }
  async function act(row: FacebookConnection, action: ConnectionAction) {
    if (!tenantID || busy || !window.confirm(t(confirmKey[action]))) return;
    setBusy(row.id);
    setError("");
    setSuccess("");
    try {
      const updated = await channels.update(tenantID, row.id, action);
      setSuccess(
        t("channels.nowStatus", {
          name: updated.account_name || updated.account_id,
          status: t(`channels.status.${updated.status}`),
        }),
      );
    } catch (err) {
      // A failed connector sync is saved on the row; repeating the action retries it.
      setError(errorText(err));
    } finally {
      setBusy("");
      setRevision((v) => v + 1);
    }
  }
  if (!tenant) return null;
  return (
    <div className="space-y-6">
      <ExpiredConnectionBanner
        tenant={tenant.id}
        connections={rows}
        owner={owner}
      />
      <Notice error={error} success={success} />
      <div className="flex flex-wrap gap-2">
        {owner && (
          <button
            className={buttonClass}
            disabled={!!busy}
            onClick={() => void connect()}
          >
            {busy === "connect" ? t("channels.facebook.opening") : t("channels.facebook.connect")}
          </button>
        )}
        <button
          className={secondaryClass}
          disabled={loading || !!busy}
          onClick={() => {
            setError("");
            setRevision((v) => v + 1);
          }}
        >
          {t("channels.refresh")}
        </button>
      </div>
      {!owner && (
        <p className="text-sm text-muted">
          {t("channels.facebook.ownerOnly")}
        </p>
      )}
      <section
        className="rounded-xl border border-border bg-surface p-5"
        aria-labelledby="facebook-pages"
      >
        <h2 id="facebook-pages" className="font-semibold">
          {t("channels.facebook.title")}
        </h2>
        {loading ? (
          <p role="status" className="mt-3">
            {t("channels.facebook.loading")}
          </p>
        ) : !rows.length ? (
          <p className="mt-3 text-sm text-muted">{t("channels.facebook.empty")}</p>
        ) : (
          <ul className="mt-2 divide-y divide-border">
            {rows.map((row) => {
              return (
                <li
                  key={row.id}
                  className="flex flex-wrap items-center justify-between gap-3 py-3"
                >
                  <div className="min-w-0">
                    <p className="font-medium">
                      {row.account_name || t("channels.facebook.fallbackName")}
                    </p>
                    <p className="break-all font-mono text-xs text-muted">
                      {row.account_id}
                    </p>
                    {row.last_error_message && (
                      <p className="mt-1 text-xs text-error">
                        {row.last_error_message}
                      </p>
                    )}
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <StatusChip tone={statusTone[row.status]}>{t(`channels.status.${row.status}`)}</StatusChip>
                    {owner && (
                      <>
                        {row.status === "auth_expired" && (
                          <button
                            className={buttonClass}
                            disabled={!!busy}
                            onClick={() => void connect()}
                          >
                            {t("channels.facebook.reconnect")}
                          </button>
                        )}
                        {(row.status === "active" ||
                          row.status === "auth_expired") && (
                          <button
                            className={secondaryClass}
                            disabled={!!busy}
                            onClick={() => void act(row, "pause")}
                          >
                            {t("channels.pause")}
                          </button>
                        )}
                        {row.status === "paused" && (
                          <button
                            className={secondaryClass}
                            disabled={!!busy}
                            onClick={() => void act(row, "resume")}
                          >
                            {t("channels.resume")}
                          </button>
                        )}
                        <button
                          className={`${secondaryClass} text-error`}
                          disabled={!!busy}
                          onClick={() => void act(row, "disconnect")}
                        >
                          {busy === row.id ? t("channels.saving") : t("channels.disconnect")}
                        </button>
                      </>
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </section>
    </div>
  );
}
