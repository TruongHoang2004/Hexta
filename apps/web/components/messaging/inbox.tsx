"use client";

import { useEffect, useRef, useState } from "react";
import { SDKApiError, type TenantMember } from "@hexta/sdk";
import { useWorkspace } from "@/components/sales/workspace";
import { sdk } from "@/lib/sdk";
import { errorText } from "@/lib/services/sales";
import {
  messaging,
  type Conversation,
  type InboxFilter,
  type InboundInput,
  type MessagingDetail,
  type SendInput,
} from "@/lib/services/messaging";
import { useAuthStore } from "@/store/useAuthStore";

const button =
  "rounded-lg border border-border px-3 py-2 text-sm disabled:opacity-50 hover:bg-background";
const inputStyle =
  "w-full rounded-lg border border-border bg-background p-2 text-sm";

function Simulator({
  tenant,
  onSaved,
}: {
  tenant: string;
  onSaved: (thread: string) => void;
}) {
  const [external, setExternal] = useState("");
  const [name, setName] = useState("");
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const pending = useRef<InboundInput | null>(null);
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (busy) return;
    const values = {
      external_key: external.trim(),
      display_name: name.trim(),
      body: body.trim(),
    };
    if (!values.external_key || !values.display_name || !values.body) return;
    if (
      !pending.current ||
      Object.entries(values).some(
        ([key, value]) =>
          pending.current?.[key as keyof InboundInput] !== value,
      )
    ) {
      pending.current = { ...values, request_key: crypto.randomUUID() };
    }
    setBusy(true);
    setError("");
    try {
      const message = await messaging.inbound(tenant, pending.current);
      pending.current = null;
      setBody("");
      onSaved(message.thread_id);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <details className="rounded-xl border border-border bg-surface p-4">
      <summary className="cursor-pointer font-medium">
        Simulate a customer message
      </summary>
      <p className="my-3 text-sm text-muted">
        Use the same customer reference to continue a conversation. This creates
        test messages only.
      </p>
      <form onSubmit={submit} className="grid gap-3">
        <label className="text-sm">
          Customer reference
          <input
            required
            maxLength={120}
            disabled={busy}
            value={external}
            onChange={(e) => setExternal(e.target.value)}
            className={inputStyle}
          />
        </label>
        <label className="text-sm">
          Customer name
          <input
            required
            maxLength={120}
            disabled={busy}
            value={name}
            onChange={(e) => setName(e.target.value)}
            className={inputStyle}
          />
        </label>
        <label className="text-sm">
          Customer message
          <textarea
            required
            maxLength={4000}
            disabled={busy}
            value={body}
            onChange={(e) => setBody(e.target.value)}
            className={inputStyle}
          />
        </label>
        {error && (
          <p role="alert" className="text-sm text-red-600">
            {error}
          </p>
        )}
        <button
          className={button}
          disabled={busy || !body.trim() || !external.trim() || !name.trim()}
        >
          {busy ? "Saving…" : "Add simulated message"}
        </button>
      </form>
    </details>
  );
}

function ThreadDetail({
  tenant,
  thread,
  owner,
  members,
  changed,
  externalRefresh,
}: {
  tenant: string;
  thread: string;
  owner: boolean;
  members: TenantMember[];
  changed: () => void;
  externalRefresh: number;
}) {
  const { user } = useAuthStore();
  const [detail, setDetail] = useState<MessagingDetail | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [body, setBody] = useState("");
  const [assignee, setAssignee] = useState("");
  const pending = useRef<SendInput | null>(null);
  const [uncertain, setUncertain] = useState(false);
  useEffect(() => {
    const controller = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const result = await messaging.detail(
          tenant,
          thread,
          "",
          controller.signal,
        );
        if (!controller.signal.aborted) {
          setDetail(result);
          setAssignee(result.conversation.turn?.assignee_id ?? "");
        }
      } catch (err) {
        if (!controller.signal.aborted) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenant, thread, refresh, externalRefresh]);
  const turn = detail?.conversation.turn;
  const assigned = !!user?.id && turn?.assignee_id === user.id;
  async function command(action: "claim" | "assign") {
    if (!turn || busy) return;
    setBusy(true);
    setError("");
    try {
      await messaging.command(tenant, turn, action, assignee || null);
      setRefresh((v) => v + 1);
      changed();
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function send(event: React.FormEvent) {
    event.preventDefault();
    if (!turn || busy || !body.trim()) return;
    const values = {
      expected_version: turn.version,
      context_revision: turn.context_revision,
      body: body.trim(),
    };
    if (!pending.current) {
      pending.current = { ...values, request_key: crypto.randomUUID() };
    }
    setBusy(true);
    setError("");
    try {
      await messaging.send(tenant, turn.id, pending.current);
      pending.current = null;
      setUncertain(false);
      setBody("");
      setRefresh((v) => v + 1);
      changed();
    } catch (err) {
      const rejected =
        err instanceof SDKApiError &&
        (err.status === 400 ||
          err.status === 409 ||
          (!uncertain && err.status >= 400 && err.status < 500));
      if (rejected) pending.current = null;
      setUncertain(!rejected);
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function older() {
    if (!detail?.next_cursor || busy) return;
    setBusy(true);
    setError("");
    try {
      const result = await messaging.detail(tenant, thread, detail.next_cursor);
      setDetail((current) =>
        current
          ? {
              ...current,
              messages: [
                ...current.messages,
                ...result.messages.filter(
                  (row) => !current.messages.some((old) => old.id === row.id),
                ),
              ],
              next_cursor: result.next_cursor,
            }
          : current,
      );
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <section
      className="min-w-0 space-y-4 rounded-xl border border-border bg-surface p-4"
      aria-label="Conversation detail"
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="font-semibold">
          {detail?.conversation.thread.display_name ?? "Conversation"}
        </h2>
        <button
          disabled={busy || loading}
          className={button}
          onClick={() => setRefresh((v) => v + 1)}
        >
          Refresh messages
        </button>
      </div>
      {error && (
        <p role="alert" className="text-sm text-red-600">
          {error} Your unsent reply is retained. Refresh and review if the
          conversation changed.
        </p>
      )}
      {uncertain && (
        <p role="status" className="text-sm text-amber-800">
          The previous save outcome is unknown. Retry the same reply to recover
          its saved result before editing or changing assignment.
        </p>
      )}
      {loading ? (
        <p role="status">Loading messages…</p>
      ) : (
        detail && (
          <>
            <div className="flex flex-wrap items-center gap-2 text-sm">
              <span className="rounded bg-background px-2 py-1">
                {turn?.assignee_id
                  ? assigned
                    ? "Assigned to you"
                    : `Assigned: ${turn.assignee_id}`
                  : "Unassigned"}
              </span>
              <span>
                {turn?.control === "sales"
                  ? "Human handling"
                  : "Awaiting human handling · AI not connected"}
              </span>
              {turn?.needs_human && <span>Needs human</span>}
              {turn?.lifecycle === "open" &&
                (!turn.assignee_id || assigned) && (
                  <button
                    disabled={busy || uncertain}
                    className={button}
                    onClick={() => void command("claim")}
                  >
                    {assigned ? "Take control" : "Claim conversation"}
                  </button>
                )}
            </div>
            {owner && turn && (
              <div className="flex flex-wrap items-end gap-2">
                <label className="text-sm">
                  Assign to
                  <select
                    className={inputStyle}
                    disabled={busy}
                    value={assignee}
                    onChange={(e) => setAssignee(e.target.value)}
                  >
                    <option value="">Unassigned queue</option>
                    {members.map((member) => (
                      <option key={member.user_id} value={member.user_id}>
                        {member.user_id === user?.id ? "You" : member.user_id}
                      </option>
                    ))}
                  </select>
                </label>
                <button
                  className={button}
                  disabled={busy || uncertain}
                  onClick={() => void command("assign")}
                >
                  Update assignment
                </button>
              </div>
            )}
            {detail.next_cursor && (
              <button
                className={button}
                disabled={busy}
                onClick={() => void older()}
              >
                Load older messages
              </button>
            )}
            <ol
              className="max-h-[480px] space-y-3 overflow-y-auto"
              aria-label="Message history"
            >
              {[...detail.messages].reverse().map((message) => (
                <li
                  key={message.id}
                  className={`max-w-[95%] rounded-xl border border-border p-3 ${message.direction === "outbound" ? "ml-auto bg-background" : "mr-auto"}`}
                >
                  <p className="text-xs text-muted">
                    {message.direction === "inbound"
                      ? "Customer · simulated input"
                      : "Sale · simulated reply"}{" "}
                    · {new Date(message.created_at).toLocaleString()}
                  </p>
                  <p className="mt-1 whitespace-pre-wrap break-words">
                    {message.body}
                  </p>
                  <p className="mt-1 text-xs text-muted">Simulation only</p>
                </li>
              ))}
            </ol>
            <form onSubmit={send} className="space-y-2">
              <label className="text-sm">
                Reply
                <textarea
                  aria-label="Reply"
                  className={inputStyle}
                  rows={3}
                  maxLength={4000}
                  value={body}
                  disabled={
                    busy || uncertain || !assigned || turn?.control !== "sales"
                  }
                  onChange={(e) => setBody(e.target.value)}
                  placeholder={
                    assigned
                      ? "Write a simulated reply…"
                      : "Claim this conversation to reply"
                  }
                />
              </label>
              <button
                className={button}
                disabled={
                  busy || !assigned || turn?.control !== "sales" || !body.trim()
                }
              >
                {uncertain ? "Retry previous save" : "Save simulated reply"}
              </button>
            </form>
          </>
        )
      )}
    </section>
  );
}

export function Inbox() {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const [filter, setFilter] = useState<InboxFilter>("all");
  const [cursor, setCursor] = useState("");
  const [next, setNext] = useState("");
  const [rows, setRows] = useState<Conversation[]>([]);
  const [selected, setSelected] = useState("");
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [refresh, setRefresh] = useState(0);
  const [detailRefresh, setDetailRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const tenantID = tenant?.id ?? "";
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const result = await messaging.list(
          tenantID,
          filter,
          cursor,
          controller.signal,
        );
        if (!controller.signal.aborted) {
          setRows(result.items);
          setNext(result.next_cursor);
        }
      } catch (err) {
        if (!controller.signal.aborted) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenantID, filter, cursor, refresh]);
  useEffect(() => {
    let ignore = false;
    if (tenantID && tenant?.owner_id === user?.id) {
      void sdk.identity
        .getUsers(tenantID)
        .then((result) => {
          if (!ignore) setMembers(result);
        })
        .catch((err) => {
          if (!ignore) setError(errorText(err));
        });
    }
    return () => {
      ignore = true;
    };
  }, [tenantID, tenant?.owner_id, user?.id]);
  if (!tenant) return null;
  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-3xl font-semibold">Inbox</h1>
        <p className="mt-2 text-sm text-muted">
          Manage customer conversations, claim responsibility and keep reply
          history.
        </p>
      </div>
      <p
        role="note"
        className="rounded-xl border border-amber-300 bg-amber-50 p-3 text-sm text-amber-950"
      >
        Simulator workspace. Messages stay inside Hexta; no customer receives a
        reply. AI, live channels and automatic turn expiry are not connected.
        Use Refresh to check new activity.
      </p>
      <Simulator
        tenant={tenant.id}
        onSaved={(thread) => {
          setFilter("all");
          setCursor("");
          setSelected(thread);
          setRefresh((v) => v + 1);
          setDetailRefresh((v) => v + 1);
        }}
      />
      <div className="grid gap-4 lg:grid-cols-[260px_minmax(0,1fr)]">
        <section
          className="space-y-3 rounded-xl border border-border bg-surface p-3"
          aria-label="Conversation list"
        >
          <label className="text-sm">
            Show
            <select
              value={filter}
              className={inputStyle}
              onChange={(e) => {
                setFilter(e.target.value as InboxFilter);
                setCursor("");
              }}
            >
              <option value="all">All conversations</option>
              <option value="unassigned">Unassigned queue</option>
              <option value="mine">Assigned to me</option>
              <option value="needs_human">Needs human</option>
            </select>
          </label>
          <button
            className={button}
            disabled={loading}
            onClick={() => setRefresh((v) => v + 1)}
          >
            Refresh inbox
          </button>
          {error && (
            <p role="alert" className="text-sm text-red-600">
              {error}
            </p>
          )}
          {loading ? (
            <p role="status">Loading conversations…</p>
          ) : (
            <>
              <ul className="space-y-2">
                {rows.map((row) => (
                  <li key={row.thread.id}>
                    <button
                      className={`w-full rounded-lg border p-3 text-left ${selected === row.thread.id ? "border-primary" : "border-border"}`}
                      onClick={() => setSelected(row.thread.id)}
                    >
                      <span className="block font-medium">
                        {row.thread.display_name}
                      </span>
                      <span className="block break-words text-xs text-muted">
                        {row.thread.external_key} ·{" "}
                        {row.turn?.assignee_id
                          ? row.turn.assignee_id === user?.id
                            ? "Yours"
                            : "Assigned"
                          : "Unassigned"}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
              {!rows.length && (
                <p className="text-sm text-muted">
                  No conversations in this view. Add a simulated customer
                  message to start.
                </p>
              )}
              <div className="flex gap-2">
                {cursor && (
                  <button className={button} onClick={() => setCursor("")}>
                    First page
                  </button>
                )}
                {next && (
                  <button className={button} onClick={() => setCursor(next)}>
                    Next page
                  </button>
                )}
              </div>
            </>
          )}
        </section>
        {selected ? (
          <ThreadDetail
            key={`${tenant.id}:${selected}`}
            tenant={tenant.id}
            thread={selected}
            owner={tenant.owner_id === user?.id}
            members={members}
            changed={() => setRefresh((v) => v + 1)}
            externalRefresh={detailRefresh}
          />
        ) : (
          <div className="rounded-xl border border-dashed border-border p-8 text-center text-muted">
            Select a conversation to see its messages.
          </div>
        )}
      </div>
    </div>
  );
}
