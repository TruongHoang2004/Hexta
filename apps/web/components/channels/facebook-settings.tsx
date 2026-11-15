"use client";

import { useEffect, useState } from "react";
import { useWorkspace } from "@/components/sales/workspace";
import { useAuthStore } from "@/store/useAuthStore";
import { errorText } from "@/lib/services/sales";
import {
  channels,
  type ConnectionAction,
  type ConnectionStatus,
  type FacebookConnection,
} from "@/lib/services/channels";
import {
  buttonClass,
  Notice,
  secondaryClass,
} from "@/components/sales/primitives";
import { ExpiredConnectionBanner } from "@/components/channels/expired-banner";

const statusLabel: Record<ConnectionStatus, [string, string]> = {
  active: ["Active", "border-green-300 bg-green-50 text-green-800"],
  paused: ["Paused", "border-amber-300 bg-amber-50 text-amber-900"],
  auth_expired: ["Reconnect required", "border-red-300 bg-red-50 text-red-800"],
  disconnected: ["Disconnected", "border-border text-muted"],
};

const confirmText: Record<ConnectionAction, string> = {
  pause:
    "Pause this Page? New messages are not received and unsent replies are cancelled until you resume.",
  resume: "Resume this Page?",
  disconnect:
    "Disconnect this Page? Its stored authorization is deleted and webhooks are unsubscribed. Conversation history is kept.",
};

export function FacebookChannelSettings() {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const owner = !!tenant && tenant.owner_id === user?.id;
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
  }, [tenantID, revision]);
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
    if (!tenantID || busy || !window.confirm(confirmText[action])) return;
    setBusy(row.id);
    setError("");
    setSuccess("");
    try {
      const updated = await channels.update(tenantID, row.id, action);
      setSuccess(
        `${updated.account_name || updated.account_id} is now ${statusLabel[updated.status][0].toLowerCase()}.`,
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
      <div>
        <h1 className="text-3xl font-semibold">Channels</h1>
        <p className="mt-2 text-sm text-muted">
          Connect Facebook Pages so customer messages arrive in the Inbox and
          assigned sales can reply. Each Page can belong to one workspace.
        </p>
      </div>
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
            {busy === "connect" ? "Opening Facebook…" : "Connect Facebook Page"}
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
          Refresh
        </button>
      </div>
      {!owner && (
        <p className="text-sm text-muted">
          Only the workspace owner can connect, pause or disconnect Pages.
        </p>
      )}
      <section
        className="rounded-xl border border-border bg-surface p-5"
        aria-labelledby="facebook-pages"
      >
        <h2 id="facebook-pages" className="font-semibold">
          Facebook Pages
        </h2>
        {loading ? (
          <p role="status" className="mt-3">
            Loading connections…
          </p>
        ) : !rows.length ? (
          <p className="mt-3 text-sm text-muted">No Pages connected yet.</p>
        ) : (
          <ul className="mt-2 divide-y divide-border">
            {rows.map((row) => {
              const [label, tone] = statusLabel[row.status];
              return (
                <li
                  key={row.id}
                  className="flex flex-wrap items-center justify-between gap-3 py-3"
                >
                  <div className="min-w-0">
                    <p className="font-medium">
                      {row.account_name || "Facebook Page"}
                    </p>
                    <p className="break-all font-mono text-xs text-muted">
                      {row.account_id}
                    </p>
                    {row.last_error_message && (
                      <p className="mt-1 text-xs text-red-700">
                        {row.last_error_message}
                      </p>
                    )}
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <span
                      className={`rounded-full border px-2 py-0.5 text-xs font-medium ${tone}`}
                    >
                      {label}
                    </span>
                    {owner && (
                      <>
                        {row.status === "auth_expired" && (
                          <button
                            className={buttonClass}
                            disabled={!!busy}
                            onClick={() => void connect()}
                          >
                            Reconnect
                          </button>
                        )}
                        {(row.status === "active" ||
                          row.status === "auth_expired") && (
                          <button
                            className={secondaryClass}
                            disabled={!!busy}
                            onClick={() => void act(row, "pause")}
                          >
                            Pause
                          </button>
                        )}
                        {row.status === "paused" && (
                          <button
                            className={secondaryClass}
                            disabled={!!busy}
                            onClick={() => void act(row, "resume")}
                          >
                            Resume
                          </button>
                        )}
                        <button
                          className={`${secondaryClass} text-red-700`}
                          disabled={!!busy}
                          onClick={() => void act(row, "disconnect")}
                        >
                          {busy === row.id ? "Saving…" : "Disconnect"}
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
