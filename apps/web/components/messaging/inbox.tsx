"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { SDKApiError, type TenantMember } from "@hexta/sdk";
import { useWorkspace } from "@/components/sales/workspace";
import { sdk } from "@/lib/sdk";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import type { Translate } from "@/lib/i18n/translate";
import {
  messaging,
  type ChannelSendInput,
  type Conversation,
  type InboxFilter,
  type InboundInput,
  type MessagingDetail,
  type MessagingMessage,
  type MessagingTurn,
  type SendInput,
} from "@/lib/services/messaging";
import {
  listAllConnections,
  platformLabel,
  type ChannelConnection,
  type ChannelPlatform,
} from "@/lib/services/channels";
import { ExpiredConnectionBanner } from "@/components/channels/expired-banner";
import {
  ConfirmDialog,
  EmptyState,
  SegmentedFilter,
  StatusChip,
  buttonClass,
  secondaryClass,
  type ChipTone,
} from "@/components/sales/primitives";
import { useAuthStore } from "@/store/useAuthStore";

const inputStyle =
  "w-full rounded-lg border border-border bg-background p-2 text-sm";
const handoffReasons = [
  "customer_requested_staff",
  "insufficient_information",
  "ai_error",
  "business_review",
  "staff_handoff",
] as const;
type HandoffReason = (typeof handoffReasons)[number];
const deliveryReasons = [
  "window_24h",
  "auth_expired",
  "rate_limit",
  "invalid_recipient",
  "bot_blocked",
  "chat_not_found",
  "transport",
  "expired",
  "connection_inactive",
  "connection_paused",
  "connection_disconnected",
] as const;
type DeliveryReason = (typeof deliveryReasons)[number];
const mediaKinds = ["image", "video", "audio", "file", "sticker"] as const;
type MediaKind = (typeof mediaKinds)[number];

function isOneOf<T extends string>(values: readonly T[], value: string): value is T {
  return (values as readonly string[]).includes(value);
}

/** Label and tone for a turn's handling state, from the viewer's perspective. */
function turnState(
  t: Translate,
  turn: MessagingTurn | null | undefined,
  userID: string | undefined,
): { label: string; tone: ChipTone } {
  switch (turn?.handling_state) {
    case "BOT_ACTIVE":
      return { label: t("desk.states.BOT_ACTIVE"), tone: "neutral" };
    case "WAITING_STAFF":
      return { label: t("desk.states.WAITING_STAFF"), tone: "warning" };
    case "STAFF_ACTIVE":
      return turn.assignee_id && turn.assignee_id === userID
        ? { label: t("desk.states.STAFF_ACTIVE_ME"), tone: "solid" }
        : { label: t("desk.states.STAFF_ACTIVE"), tone: "neutral" };
    default:
      return { label: t("desk.states.UNKNOWN"), tone: "warning" };
  }
}

function MessageContent({
  message,
  platform,
}: {
  message: MessagingMessage;
  platform: ChannelPlatform | null;
}) {
  const t = useT();
  if (message.content_type === "unsupported_media") {
    const kinds = message.placeholder_meta?.media_kinds ?? [];
    const names = kinds.length
      ? kinds
          .map((k) => (isOneOf<MediaKind>(mediaKinds, k) ? t(`desk.media.${k}`) : t("desk.media.attachment")))
          .join(", ")
      : t("desk.media.attachment");
    return (
      <p className="mt-1 rounded-lg border border-dashed border-border p-2 text-sm text-muted">
        {t("desk.media.placeholder", {
          kinds: names,
          channel: platform ? platformLabel[platform] : t("desk.media.theChannel"),
        })}
      </p>
    );
  }
  return <p className="mt-1 whitespace-pre-wrap break-words">{message.body}</p>;
}

function DeliveryStatus({
  message,
  platform,
}: {
  message: MessagingMessage;
  platform: ChannelPlatform | null;
}) {
  const t = useT();
  if (message.status === "simulated" || !platform)
    return <p className="mt-1 text-xs text-muted">{t("desk.delivery.simulated")}</p>;
  const channel = platformLabel[platform];
  if (message.status === "received")
    return <p className="mt-1 text-xs text-muted">{t("desk.delivery.received", { channel })}</p>;
  switch (message.delivery_status) {
    case "sent":
      return (
        <p className="mt-1 text-xs">
          <StatusChip tone="success">{t("desk.delivery.sent", { channel })}</StatusChip>
        </p>
      );
    case "failed": {
      const reason = message.delivery_error_reason;
      return (
        <p role="note" className="mt-1 text-xs text-error">
          {isOneOf<DeliveryReason>(deliveryReasons, reason)
            ? t(`desk.delivery.reasons.${reason}`, { channel })
            : t("desk.delivery.failed", { channel })}
        </p>
      );
    }
    case "unknown":
      return (
        <div role="note" className="mt-1 space-y-1 text-xs">
          <StatusChip tone="warning">{t("desk.delivery.unknownChip")}</StatusChip>
          <p>
            {t("desk.delivery.unknown", {
              channel,
              place:
                platform === "telegram"
                  ? t("desk.delivery.placeTelegram")
                  : t("desk.delivery.placeFacebook"),
            })}
          </p>
        </div>
      );
    default:
      return <p className="mt-1 text-xs text-muted">{t("desk.delivery.sending")}</p>;
  }
}

function Simulator({
  tenant,
  onSaved,
}: {
  tenant: string;
  onSaved: (thread: string) => void;
}) {
  const t = useT();
  const errorText = useErrorText();
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
      <summary className="cursor-pointer font-medium">{t("desk.simulator.title")}</summary>
      <p className="my-3 text-sm text-muted">{t("desk.simulator.help")}</p>
      <form onSubmit={submit} className="grid gap-3 md:grid-cols-2">
        <label className="text-sm">
          {t("desk.simulator.reference")}
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
          {t("desk.simulator.name")}
          <input
            required
            maxLength={120}
            disabled={busy}
            value={name}
            onChange={(e) => setName(e.target.value)}
            className={inputStyle}
          />
        </label>
        <label className="text-sm md:col-span-2">
          {t("desk.simulator.message")}
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
          <p role="alert" className="text-sm text-error md:col-span-2">
            {error}
          </p>
        )}
        <div className="md:col-span-2">
          <button
            className={secondaryClass}
            disabled={busy || !body.trim() || !external.trim() || !name.trim()}
          >
            {busy ? t("desk.simulator.saving") : t("desk.simulator.submit")}
          </button>
        </div>
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
  connections,
}: {
  tenant: string;
  thread: string;
  owner: boolean;
  members: TenantMember[];
  changed: () => void;
  externalRefresh: number;
  // null while unknown; the API still rejects sends on inactive connections.
  connections: ChannelConnection[] | null;
}) {
  const t = useT();
  const errorText = useErrorText();
  const { user } = useAuthStore();
  const [detail, setDetail] = useState<MessagingDetail | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [body, setBody] = useState("");
  const [assignee, setAssignee] = useState("");
  const [reason, setReason] = useState<HandoffReason>("staff_handoff");
  const [handoffOpen, setHandoffOpen] = useState(false);
  const pending = useRef<SendInput | null>(null);
  const pendingChannel = useRef<ChannelSendInput | null>(null);
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
  }, [tenant, thread, refresh, externalRefresh, errorText]);
  const turn = detail?.conversation.turn;
  const assigned = !!user?.id && turn?.assignee_id === user.id;
  async function command(action: "claim" | "assign" | "handoff" | "return-to-bot") {
    if (!turn || busy) return;
    setBusy(true);
    setError("");
    try {
      await messaging.command(tenant, turn, action, assignee || null, reason);
      setHandoffOpen(false);
      setRefresh((v) => v + 1);
      changed();
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  const source = detail?.conversation.source;
  const platform: ChannelPlatform | null =
    source === "facebook" || source === "telegram" ? source : null;
  const channelName = platform ? platformLabel[platform] : "";
  // A known connection that is not active (paused, auth_expired, or disconnected and so absent) locks the composer.
  const connectionID = detail?.conversation.thread.connection_id;
  const connection = connections?.find((c) => c.id === connectionID);
  const locked =
    !!platform &&
    !!connectionID &&
    connections !== null &&
    connection?.status !== "active";
  const lockReason = !connection
    ? t("desk.readOnly.disconnected", { channel: channelName })
    : connection.status === "paused"
      ? t("desk.readOnly.paused", { channel: channelName })
      : platform === "telegram"
        ? t("desk.readOnly.expiredTelegram")
        : t("desk.readOnly.expiredFacebook");
  const canReply = !locked && assigned && turn?.control === "sales";
  const readOnlyReason = locked
    ? lockReason
    : canReply || !turn
      ? ""
      : turn.assignee_id && !assigned
        ? t("desk.readOnly.otherStaff")
        : turn.handling_state === "BOT_ACTIVE"
          ? t("desk.readOnly.bot")
          : t("desk.readOnly.waiting");
  async function send(event: React.FormEvent) {
    event.preventDefault();
    if (!turn || busy || !body.trim() || locked) return;
    if (platform) return sendChannel(turn.id, turn.version, turn.context_revision);
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
  // Real sends keep the same client_message_id across uncertain retries so the API replays instead of duplicating.
  async function sendChannel(turnID: string, version: number, context: number) {
    if (!pendingChannel.current) {
      pendingChannel.current = {
        client_message_id: crypto.randomUUID(),
        text: body.trim(),
        expected_version: version,
        context_revision: context,
      };
    }
    setBusy(true);
    setError("");
    try {
      await messaging.sendChannel(tenant, turnID, pendingChannel.current);
      pendingChannel.current = null;
      setUncertain(false);
      setBody("");
      setRefresh((v) => v + 1);
      changed();
    } catch (err) {
      const rejected =
        err instanceof SDKApiError && err.status >= 400 && err.status < 500;
      if (rejected) pendingChannel.current = null;
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
  const state = turnState(t, turn, user?.id);
  const sourceLabel = source ? t(`desk.sources.${source}`) : "";
  const canHandoff =
    turn?.lifecycle === "open" &&
    turn.handling_state !== "WAITING_STAFF" &&
    (assigned || owner || turn.handling_state === "BOT_ACTIVE");
  const canReturn = turn?.lifecycle === "open" && (assigned || owner) && turn.control === "sales";
  const canClaim = turn?.lifecycle === "open" && (!turn.assignee_id || assigned);
  return (
    <>
      <section
        className="flex min-w-0 flex-[999_1_420px] flex-col rounded-xl border border-border bg-surface"
        aria-label={t("desk.detailLabel")}
      >
        <header className="flex flex-wrap items-center justify-between gap-3 border-b border-border p-4">
          <div className="min-w-0">
            <h2 className="truncate text-lg font-semibold">
              {detail?.conversation.thread.display_name ?? t("desk.conversation")}
            </h2>
            {turn && (
              <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted">
                <StatusChip tone={state.tone}>{state.label}</StatusChip>
                {turn.needs_human && <StatusChip tone="warning">{t("desk.needsHuman")}</StatusChip>}
                <span>{sourceLabel}</span>
              </div>
            )}
          </div>
          <div className="flex flex-wrap gap-2">
            {canClaim && (
              <button
                type="button"
                disabled={busy || uncertain}
                className={buttonClass}
                onClick={() => void command("claim")}
              >
                {assigned ? t("desk.actions.takeControl") : t("desk.actions.claim")}
              </button>
            )}
            {canHandoff && (
              <button
                type="button"
                className={secondaryClass}
                disabled={busy || uncertain}
                onClick={() => setHandoffOpen(true)}
              >
                {t("desk.actions.handoff")}
              </button>
            )}
            {canReturn && (
              <button
                type="button"
                className={secondaryClass}
                disabled={busy || uncertain}
                onClick={() => void command("return-to-bot")}
              >
                {t("desk.actions.returnToBot")}
              </button>
            )}
            <button
              type="button"
              disabled={busy || loading}
              className={secondaryClass}
              onClick={() => setRefresh((v) => v + 1)}
            >
              {t("desk.refresh")}
            </button>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-3 bg-background/60 p-4">
          {error && (
            <p role="alert" className="text-sm text-error">
              {error} {t("desk.composer.errorSuffix")}
            </p>
          )}
          {uncertain && (
            <p role="status" className="rounded-lg border border-warning bg-warning/15 p-2 text-sm">
              {t("desk.composer.uncertain")}
            </p>
          )}
          {loading ? (
            <p role="status">{t("desk.loadingMessages")}</p>
          ) : (
            detail && (
              <>
                {detail.next_cursor && (
                  <button
                    type="button"
                    className={`${secondaryClass} self-center`}
                    disabled={busy}
                    onClick={() => void older()}
                  >
                    {t("desk.loadOlder")}
                  </button>
                )}
                <ol
                  className="max-h-[520px] space-y-3 overflow-y-auto"
                  aria-label={t("desk.history")}
                >
                  {[...detail.messages].reverse().map((message) => (
                    <li
                      key={message.id}
                      className={`max-w-[85%] rounded-xl border border-border p-3 ${message.direction === "outbound" ? "ml-auto bg-primary/5" : "mr-auto bg-surface"}`}
                    >
                      <p className="text-xs text-muted">
                        {message.status === "received"
                          ? `${t("desk.sender.customer")} · ${detail.conversation.thread.display_name}`
                          : message.status === "channel"
                            ? t("desk.sender.sale", { channel: channelName })
                            : message.direction === "inbound"
                              ? t("desk.sender.customerSimulated")
                              : t("desk.sender.saleSimulated")}{" "}
                        · {new Date(message.created_at).toLocaleString()}
                      </p>
                      <MessageContent message={message} platform={platform} />
                      <DeliveryStatus message={message} platform={platform} />
                    </li>
                  ))}
                </ol>
              </>
            )
          )}
        </div>
        {detail && (
          <form onSubmit={send} className="space-y-2 border-t border-border p-4">
            {readOnlyReason && (
              <p role="status" className="rounded-lg border border-border bg-background p-2 text-sm">
                {readOnlyReason}
              </p>
            )}
            <label className="block text-sm">
              <span className="sr-only">{t("desk.composer.label")}</span>
              <textarea
                className={inputStyle}
                rows={3}
                maxLength={platform ? 2000 : 4000}
                value={body}
                disabled={busy || uncertain || !canReply}
                onChange={(e) => setBody(e.target.value)}
                placeholder={
                  !canReply
                    ? t("desk.composer.placeholderLocked")
                    : platform
                      ? t("desk.composer.placeholderChannel", { channel: channelName })
                      : t("desk.composer.placeholderSimulated")
                }
              />
            </label>
            <div className="flex justify-end">
              <button className={buttonClass} disabled={busy || !canReply || !body.trim()}>
                {uncertain
                  ? platform
                    ? t("desk.composer.retrySend")
                    : t("desk.composer.retrySave")
                  : platform
                    ? t("desk.composer.send", { channel: channelName })
                    : t("desk.composer.saveSimulated")}
              </button>
            </div>
          </form>
        )}
      </section>
      <aside
        aria-label={t("desk.customer.title")}
        className="flex min-w-0 flex-[1_1_240px] flex-col gap-3 self-start rounded-xl border border-border bg-surface p-4"
      >
        <h2 className="font-semibold">{t("desk.customer.title")}</h2>
        {detail && (
          <dl className="space-y-1 text-sm">
            <dt className="sr-only">{t("desk.conversation")}</dt>
            <dd className="font-medium">{detail.conversation.thread.display_name}</dd>
            <dt className="text-muted">{t("desk.customer.reference")}</dt>
            <dd className="break-words">{detail.conversation.thread.external_key}</dd>
            <dd>
              <StatusChip>{sourceLabel}</StatusChip>
            </dd>
          </dl>
        )}
        <Link href={`/tenants/${tenant}/orders`} className={`${buttonClass} text-center`} aria-describedby="draft-hint">
          {t("desk.customer.createDraft")}
        </Link>
        <p id="draft-hint" className="text-xs text-muted">
          {t("desk.customer.draftHint")}
        </p>
        <Link href={`/tenants/${tenant}/storefront`} className={`${secondaryClass} text-center`}>
          {t("desk.customer.askConsultant")}
        </Link>
        {owner && turn && (
          <div className="space-y-2 border-t border-border pt-3">
            <label className="block text-sm">
              {t("desk.actions.assignTo")}
              <select
                className={inputStyle}
                disabled={busy}
                value={assignee}
                onChange={(e) => setAssignee(e.target.value)}
              >
                <option value="">{t("desk.actions.assignQueue")}</option>
                {members.map((member) => (
                  <option key={member.user_id} value={member.user_id}>
                    {member.user_id === user?.id ? t("desk.actions.you") : member.user_id}
                  </option>
                ))}
              </select>
            </label>
            <button
              type="button"
              className={secondaryClass}
              disabled={busy || uncertain}
              onClick={() => void command("assign")}
            >
              {t("desk.actions.assign")}
            </button>
          </div>
        )}
      </aside>
      <ConfirmDialog
        open={handoffOpen}
        title={t("desk.handoffDialog.title")}
        confirmLabel={t("desk.handoffDialog.confirm")}
        cancelLabel={t("common.cancel")}
        busy={busy}
        onCancel={() => setHandoffOpen(false)}
        onConfirm={() => void command("handoff")}
      >
        <p>{t("desk.handoffDialog.body")}</p>
        <fieldset className="space-y-1">
          <legend className="mb-1 font-medium">{t("desk.handoffDialog.reasonLegend")}</legend>
          {handoffReasons.map((code) => (
            <label key={code} className="flex min-h-9 items-center gap-2">
              <input
                type="radio"
                name="handoff-reason"
                value={code}
                checked={reason === code}
                onChange={() => setReason(code)}
              />
              {t(`desk.reasons.${code}`)}
            </label>
          ))}
        </fieldset>
      </ConfirmDialog>
    </>
  );
}

type DeskFilter = Extract<InboxFilter, "bot" | "needs_human" | "mine" | "all">;

export function Inbox() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const errorText = useErrorText();
  const { user } = useAuthStore();
  const [filter, setFilter] = useState<DeskFilter>("all");
  const [cursor, setCursor] = useState("");
  const [next, setNext] = useState("");
  const [rows, setRows] = useState<Conversation[]>([]);
  const [selected, setSelected] = useState("");
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [refresh, setRefresh] = useState(0);
  const [detailRefresh, setDetailRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [connections, setConnections] = useState<ChannelConnection[] | null>(
    null,
  );
  const tenantID = tenant?.id ?? "";
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    // Banner and composer-lock data only; a failure here must not hide the Inbox.
    listAllConnections(tenantID, controller.signal)
      .then((rows) => {
        if (!controller.signal.aborted) setConnections(rows);
      })
      .catch(() => undefined);
    return () => controller.abort();
  }, [tenantID, refresh]);
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
  }, [tenantID, filter, cursor, refresh, errorText]);
  useEffect(() => {
    let ignore = false;
    if (tenantID && isOwner) {
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
  }, [tenantID, isOwner, errorText]);
  if (!tenant) return null;
  const filters: { value: DeskFilter; label: string }[] = [
    { value: "bot", label: t("desk.filters.bot") },
    { value: "needs_human", label: t("desk.filters.needs_human") },
    { value: "mine", label: t("desk.filters.mine") },
    { value: "all", label: t("desk.filters.all") },
  ];
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("desk.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("desk.subtitle")}</p>
        </div>
        <button
          type="button"
          className={secondaryClass}
          disabled={loading}
          onClick={() => setRefresh((v) => v + 1)}
        >
          {t("desk.refresh")}
        </button>
      </div>
      <p role="note" className="rounded-lg border border-border bg-surface p-3 text-sm text-muted">
        {t("desk.notice")}
      </p>
      <ExpiredConnectionBanner
        tenant={tenant.id}
        connections={connections ?? []}
        owner={isOwner}
      />
      <div className="flex flex-wrap items-start gap-4">
        <section
          className="flex min-w-0 flex-[1_1_280px] flex-col gap-3 rounded-xl border border-border bg-surface p-3 lg:max-w-[340px]"
          aria-label={t("desk.listLabel")}
        >
          <SegmentedFilter
            label={t("desk.filterLabel")}
            options={filters}
            value={filter}
            onChange={(value) => {
              setFilter(value);
              setCursor("");
            }}
          />
          {error && (
            <p role="alert" className="text-sm text-error">
              {error}
            </p>
          )}
          {loading ? (
            <p role="status">{t("desk.loadingList")}</p>
          ) : rows.length ? (
            <ul className="-mx-3 divide-y divide-border">
              {rows.map((row) => {
                const state = turnState(t, row.turn, user?.id);
                const active = selected === row.thread.id;
                return (
                  <li key={row.thread.id}>
                    <button
                      type="button"
                      aria-current={active ? "true" : undefined}
                      className={`flex w-full flex-col gap-1 px-3 py-3 text-left hover:bg-background ${active ? "bg-primary/10 shadow-[inset_3px_0_0_var(--primary)]" : ""}`}
                      onClick={() => setSelected(row.thread.id)}
                    >
                      <span className="flex items-center justify-between gap-2">
                        <span className="truncate font-medium">{row.thread.display_name}</span>
                        <span className="text-xs text-muted">{t(`desk.sources.${row.source}`)}</span>
                      </span>
                      <span className="flex flex-wrap items-center gap-1.5">
                        <StatusChip tone={state.tone}>{state.label}</StatusChip>
                        {row.turn?.needs_human && row.turn.handling_state !== "WAITING_STAFF" && (
                          <StatusChip tone="warning">{t("desk.needsHuman")}</StatusChip>
                        )}
                      </span>
                      <span className="text-xs text-muted">
                        {row.turn?.assignee_id
                          ? row.turn.assignee_id === user?.id
                            ? t("desk.assignedYou")
                            : t("desk.assignedOther")
                          : t("desk.unassigned")}
                        {row.turn && ` · ${new Date(row.turn.last_activity_at).toLocaleString()}`}
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          ) : (
            <EmptyState
              title={t("desk.emptyTitle")}
              description={t("desk.emptyBody")}
              action={
                isOwner ? (
                  <Link href={`/tenants/${tenant.id}/settings/channels`} className={secondaryClass}>
                    {t("desk.connectChannels")}
                  </Link>
                ) : undefined
              }
            />
          )}
          <div className="flex gap-2">
            {cursor && (
              <button type="button" className={secondaryClass} onClick={() => setCursor("")}>
                {t("desk.firstPage")}
              </button>
            )}
            {next && (
              <button type="button" className={secondaryClass} onClick={() => setCursor(next)}>
                {t("desk.nextPage")}
              </button>
            )}
          </div>
        </section>
        {selected ? (
          <ThreadDetail
            key={`${tenant.id}:${selected}`}
            tenant={tenant.id}
            thread={selected}
            owner={isOwner}
            members={members}
            changed={() => setRefresh((v) => v + 1)}
            externalRefresh={detailRefresh}
            connections={connections}
          />
        ) : (
          <div className="flex-[999_1_420px] rounded-xl border border-dashed border-border p-8 text-center text-muted">
            {t("desk.selectPrompt")}
          </div>
        )}
      </div>
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
    </div>
  );
}
