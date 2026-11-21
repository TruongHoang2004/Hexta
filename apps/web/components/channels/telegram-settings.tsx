"use client";

import { useEffect, useState } from "react";
import { useWorkspace } from "@/components/sales/workspace";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
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
  StatusChip,
  secondaryClass,
} from "@/components/sales/primitives";
import { statusTone } from "@/components/channels/facebook-settings";
import { ExpiredConnectionBanner } from "@/components/channels/expired-banner";

const confirmKey = {
  pause: "channels.telegram.confirmPause",
  resume: "channels.telegram.confirmResume",
  disconnect: "channels.telegram.confirmDisconnect",
} as const satisfies Record<Exclude<TelegramAction, "rotate_token">, string>;

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
  const { tenant, isOwner: owner } = useWorkspace();
  const t = useT();
  const errorText = useErrorText();
  const statusText = (status: ConnectionStatus) =>
    status === "auth_expired" ? t("channels.telegram.tokenRejected") : t(`channels.status.${status}`);
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
  }, [tenantID, revision, errorText]);
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
      setSuccess(t("channels.telegram.connected", { name: botName(created) }));
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
      setSuccess(t("channels.telegram.rotated", { name: botName(updated) }));
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
    if (!tenantID || busy || !window.confirm(t(confirmKey[action]))) return;
    setBusy(row.id);
    setError("");
    setSuccess("");
    try {
      const updated = await telegram.update(tenantID, row.id, action);
      setSuccess(
        t("channels.nowStatus", { name: botName(updated), status: statusText(updated.status) }),
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
          {t("channels.telegram.title")}
        </h2>
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
      <p className="text-sm text-muted">
        {t("channels.telegram.intro")}
      </p>
      <ExpiredConnectionBanner
        tenant={tenant.id}
        connections={rows}
        owner={owner}
      />
      <Notice error={error} success={success} />
      {owner ? (
        <form onSubmit={connect} className="space-y-2">
          <Field title={t("channels.telegram.token")}>
            <TokenInput
              id="telegram-token"
              value={token}
              disabled={!!busy}
              onChange={setToken}
            />
          </Field>
          <p id="telegram-token-help" className="text-xs text-muted">
            {t("channels.telegram.tokenHelp")}
          </p>
          <button
            className={buttonClass}
            disabled={!!busy || !tokenPattern.test(token)}
          >
            {busy === "connect" ? t("channels.telegram.connecting") : t("channels.telegram.connect")}
          </button>
        </form>
      ) : (
        <p className="text-sm text-muted">
          {t("channels.telegram.ownerOnly")}
        </p>
      )}
      {loading ? (
        <p role="status">{t("channels.telegram.loading")}</p>
      ) : !rows.length ? (
        <p className="text-sm text-muted">{t("channels.telegram.empty")}</p>
      ) : (
        <ul className="divide-y divide-border">
          {rows.map((row) => {
            return (
              <li key={row.id} className="space-y-2 py-3">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="min-w-0">
                    <p className="font-medium">
                      {row.account_name || t("channels.telegram.fallbackName")}
                    </p>
                    <p className="break-all font-mono text-xs text-muted">
                      {row.account_username ? `@${row.account_username} · ` : ""}
                      {t("channels.telegram.id", { id: row.account_id })}
                    </p>
                    {row.last_error_message && (
                      <p className="mt-1 text-xs text-error">
                        {row.last_error_message}
                      </p>
                    )}
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <StatusChip tone={statusTone[row.status]}>{statusText(row.status)}</StatusChip>
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
                          {t("channels.telegram.rotate")}
                        </button>
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
                </div>
                {owner && rotating === row.id && (
                  <form
                    onSubmit={(e) => void rotate(e, row)}
                    className="space-y-2 rounded-lg border border-border p-3"
                  >
                    <Field title={t("channels.telegram.newToken", { name: botName(row) })}>
                      <TokenInput
                        id={`rotate-${row.id}`}
                        value={rotateToken}
                        disabled={!!busy}
                        onChange={setRotateToken}
                      />
                    </Field>
                    <p id={`rotate-${row.id}-help`} className="text-xs text-muted">
                      {t("channels.telegram.rotateHelp")}
                    </p>
                    <div className="flex gap-2">
                      <button
                        className={buttonClass}
                        disabled={!!busy || !tokenPattern.test(rotateToken)}
                      >
                        {busy === row.id ? t("channels.telegram.rotating") : t("channels.telegram.saveToken")}
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
                        {t("channels.cancel")}
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
