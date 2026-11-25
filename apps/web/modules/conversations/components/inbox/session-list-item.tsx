"use client";

import Link from "next/link";
import { StatusChip, secondaryClass } from "@/components/primitives";
import { useLocale, useT } from "@/lib/i18n/provider";
import type { MessagingSession } from "../../services/messaging";
import { sessionCause, sessionOutcome, sessionRange } from "../../utils";

/** One session in the "Phiên" panel: time range, state, handler, counts, orders and its actions. */
export function SessionListItem({
  tenantID,
  session,
  current,
  handlerName,
  canClose,
  canEditOutcome,
  disabled,
  onJump,
  onClose,
  onEditOutcome,
}: {
  tenantID: string;
  session: MessagingSession;
  current: boolean;
  handlerName: string;
  canClose: boolean;
  canEditOutcome: boolean;
  disabled: boolean;
  onJump: () => void;
  onClose: () => void;
  onEditOutcome: () => void;
}) {
  const t = useT();
  const locale = useLocale();
  const { turn } = session;
  const open = turn.lifecycle === "open";
  return (
    <li
      aria-current={current ? "true" : undefined}
      className={`space-y-1.5 rounded-lg border p-2.5 text-sm ${current ? "border-primary bg-primary/5" : "border-border"}`}
    >
      <button
        type="button"
        className="w-full text-left font-medium underline-offset-2 hover:underline focus-visible:underline"
        title={t("desk.sessions.jump")}
        onClick={onJump}
      >
        {sessionRange(t, session, locale)}
      </button>
      <div className="flex flex-wrap items-center gap-1.5 text-xs">
        {current && <StatusChip tone="solid">{t("desk.sessions.current")}</StatusChip>}
        {open ? (
          <StatusChip tone="warning">{t("desk.sessions.open")}</StatusChip>
        ) : (
          <StatusChip>{sessionCause(t, turn)}</StatusChip>
        )}
        {!open && <span>{sessionOutcome(t, turn)}</span>}
      </div>
      <p className="text-xs text-muted">
        {handlerName ? t("desk.sessions.handledBy", { name: handlerName }) : t("desk.sessions.noHandler")} ·{" "}
        {t("desk.sessions.messages", { count: session.message_count })}
      </p>
      {turn.note && <p className="whitespace-pre-wrap break-words text-xs">{turn.note}</p>}
      {session.orders.length > 0 && (
        <p className="text-xs">
          {t("desk.sessions.orders")}{" "}
          {session.orders.map((order, index) => (
            <span key={order.id}>
              {index > 0 && ", "}
              <Link className="underline" href={`/tenants/${tenantID}/orders?order=${order.id}`}>
                {order.order_number}
              </Link>
            </span>
          ))}
        </p>
      )}
      {(canClose || canEditOutcome) && (
        <button
          type="button"
          className={`${secondaryClass} w-full`}
          disabled={disabled}
          onClick={canClose ? onClose : onEditOutcome}
        >
          {canClose ? t("desk.sessions.closeAction") : t("desk.sessions.editOutcome")}
        </button>
      )}
    </li>
  );
}
