"use client";

import { useEffect, useState } from "react";
import { useWorkspace } from "@/components/sales/workspace";
import { useAuthStore } from "@/store/useAuthStore";
import { errorText } from "@/lib/services/sales";
import {
  telegram,
  type ConnectionStatus,
  type TelegramAction,
  type TelegramConnection,
} from "@/lib/services/channels";
import {
  buttonClass,
  Field,
  inputClass,
  Notice,
  secondaryClass,
} from "@/components/sales/primitives";
import { ExpiredConnectionBanner } from "@/components/channels/expired-banner";

const statusLabel: Record<ConnectionStatus, [string, string]> = {
  active: ["Active", "border-green-300 bg-green-50 text-green-800"],
  paused: ["Paused", "border-amber-300 bg-amber-50 text-amber-900"],
  auth_expired: ["Token rejected", "border-red-300 bg-red-50 text-red-800"],
  disconnected: ["Disconnected", "border-border text-muted"],
};

const confirmText: Record<Exclude<TelegramAction, "rotate_token">, string> = {
  pause:
    "Pause this bot? Telegram stops delivering messages, messages sent while paused are not received, and unsent replies are cancelled until you resume.",
  resume: "Resume this bot?",
  disconnect:
    "Disconnect this bot? Its stored Bot Token is deleted and the webhook is removed. Conversation history is kept.",
};

// Mirrors the API check: numeric bot ID, colon, secret from @BotFather.
const tokenPattern = /^[0-9]{1,20}:[A-Za-z0-9_-]{30,80}$/;

function TokenInput({
  id,
  value,
  disabled,
  onChange,
}: {
  id: string;
  value: string;
  disabled: boolean;
  onChange: (value: string) => void;
}) {
  return (
    <input
      id={id}
      type="password"
      autoComplete="off"
      spellCheck={false}
      required
      maxLength={128}
      pattern="[0-9]{1,20}:[A-Za-z0-9_\-]{30,80}"
      placeholder="123456789:AA…"
      aria-describedby={`${id}-help`}
      className={inputClass}
      disabled={disabled}
      value={value}
      onChange={(e) => onChange(e.target.value.trim())}
    />
  );
}

export function TelegramChannelSettings() {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const owner = !!tenant && tenant.owner_id === user?.id;
  const [rows, setRows] = useState<TelegramConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [revision, setRevision] = useState(0);
  // Raw tokens live only until submission and are cleared whatever the outcome.
  const [token, setToken] = useState("");
  const [rotating, setRotating] = useState("");
  const [rotateToken, setRotateToken] = useState("");
  const tenantID = tenant?.id ?? "";
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const result = await telegram.list(tenantID, controller.signal);
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
  function botName(row: TelegramConnection) {
    return row.account_username
      ? `@${row.account_username}`
      : row.account_name || row.account_id;
  }
  async function connect(event: React.FormEvent) {
    event.preventDefault();
    if (!tenantID || busy || !tokenPattern.test(token)) return;
    const submitted = token;
    setToken("");
    setBusy("connect");
    setError("");
    setSuccess("");
    try {
      const created = await telegram.connect(tenantID, submitted);
      setSuccess(`${botName(created)} is connected.`);
    } catch (err) {
      // A failed connector sync is saved on the row; connecting the same token again retries it.
      setError(errorText(err));
    } finally {
      setBusy("");
      setRevision((v) => v + 1);
    }
  }
  async function rotate(event: React.FormEvent, row: TelegramConnection) {
    event.preventDefault();
    if (!tenantID || busy || !tokenPattern.test(rotateToken)) return;
    const submitted = rotateToken;
    setRotateToken("");
    setBusy(row.id);
    setError("");
    setSuccess("");
    try {
      const updated = await telegram.update(
        tenantID,
        row.id,
        "rotate_token",
        submitted,
      );
      setRotating("");
      setSuccess(`Token rotated for ${botName(updated)}.`);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy("");
      setRevision((v) => v + 1);
    }
  }
  async function act(
    row: TelegramConnection,
    action: Exclude<TelegramAction, "rotate_token">,
  ) {
    if (!tenantID || busy || !window.confirm(confirmText[action])) return;
    setBusy(row.id);
    setError("");
    setSuccess("");
    try {
      const updated = await telegram.update(tenantID, row.id, action);
      setSuccess(
        `${botName(updated)} is now ${statusLabel[updated.status][0].toLowerCase()}.`,
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
    <section
      className="space-y-4 rounded-xl border border-border bg-surface p-5"
      aria-labelledby="telegram-bots"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 id="telegram-bots" className="font-semibold">
          Telegram bots
        </h2>
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
      <p className="text-sm text-muted">
        Customers message your bot in a 1-1 Telegram chat. Group and channel
        messages are ignored. Each bot can belong to one workspace.
      </p>
      <ExpiredConnectionBanner
        tenant={tenant.id}
        connections={rows}
        owner={owner}
      />
      <Notice error={error} success={success} />
      {owner ? (
        <form onSubmit={connect} className="space-y-2">
          <Field title="Bot Token">
            <TokenInput
              id="telegram-token"
              value={token}
              disabled={!!busy}
              onChange={setToken}
            />
          </Field>
          <p id="telegram-token-help" className="text-xs text-muted">
            Create a bot with @BotFather and paste its token. The token is
            stored encrypted and is never shown again.
          </p>
          <button
            className={buttonClass}
            disabled={!!busy || !tokenPattern.test(token)}
          >
            {busy === "connect" ? "Connecting…" : "Connect Telegram bot"}
          </button>
        </form>
      ) : (
        <p className="text-sm text-muted">
          Only the workspace owner can connect, pause, rotate or disconnect
          bots.
        </p>
      )}
      {loading ? (
        <p role="status">Loading bots…</p>
      ) : !rows.length ? (
        <p className="text-sm text-muted">No bots connected yet.</p>
      ) : (
        <ul className="divide-y divide-border">
          {rows.map((row) => {
            const [label, tone] = statusLabel[row.status];
            return (
              <li key={row.id} className="space-y-2 py-3">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="min-w-0">
                    <p className="font-medium">
                      {row.account_name || "Telegram bot"}
                    </p>
                    <p className="break-all font-mono text-xs text-muted">
                      {row.account_username ? `@${row.account_username} · ` : ""}
                      ID {row.account_id}
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
                        <button
                          className={
                            row.status === "auth_expired"
                              ? buttonClass
                              : secondaryClass
                          }
                          disabled={!!busy}
                          aria-expanded={rotating === row.id}
                          onClick={() => {
                            setRotateToken("");
                            setRotating(rotating === row.id ? "" : row.id);
                          }}
                        >
                          Rotate token
                        </button>
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
                </div>
                {owner && rotating === row.id && (
                  <form
                    onSubmit={(e) => void rotate(e, row)}
                    className="space-y-2 rounded-lg border border-border p-3"
                  >
                    <Field title={`New Bot Token for ${botName(row)}`}>
                      <TokenInput
                        id={`rotate-${row.id}`}
                        value={rotateToken}
                        disabled={!!busy}
                        onChange={setRotateToken}
                      />
                    </Field>
                    <p id={`rotate-${row.id}-help`} className="text-xs text-muted">
                      Use a token for this same bot (for example after
                      /revoke in @BotFather). Rotating re-activates the bot. A
                      different bot must be connected separately.
                    </p>
                    <div className="flex gap-2">
                      <button
                        className={buttonClass}
                        disabled={!!busy || !tokenPattern.test(rotateToken)}
                      >
                        {busy === row.id ? "Rotating…" : "Save new token"}
                      </button>
                      <button
                        type="button"
                        className={secondaryClass}
                        disabled={!!busy}
                        onClick={() => {
                          setRotateToken("");
                          setRotating("");
                        }}
                      >
                        Cancel
                      </button>
                    </div>
                  </form>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
