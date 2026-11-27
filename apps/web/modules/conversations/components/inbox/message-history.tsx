"use client";

import { Fragment, useEffect, useEffectEvent, useRef, useState } from "react";
import { secondaryClass } from "@/components/primitives";
import { useLocale, useT } from "@/lib/i18n/provider";
import type { ChannelPlatform } from "@/modules/settings";
import type { MessagingDetail, MessagingSession } from "../../services/messaging";
import type { SessionFocus } from "../../hooks/use-session-jump";
import { sessionCause, sessionOutcome, sessionRange } from "../../utils";
import { MessageBubble } from "./message-bubble";
import { TurnDivider } from "./turn-divider";

const NEAR_BOTTOM_PX = 80;

/** DOM id of the divider that starts a session in the history. */
export const sessionAnchorID = (turnID: string) => `session-start-${turnID}`;

/**
 * Messages oldest-first, with a button to page in older ones. A divider starts each session (turn) with its time
 * range and outcome, and marks the end of a closed current session. A newly arrived message scrolls into view when
 * the reader is already at the bottom; otherwise a "new messages" button keeps their place. `focus` scrolls to a
 * session's start once its first message is loaded, then `onFocused` clears it so later loads (polls, older pages)
 * do not scroll back.
 */
export function MessageHistory({
  detail,
  sessions,
  platform,
  busy,
  focus,
  onFocused,
  onLoadOlder,
}: {
  detail: MessagingDetail;
  /** Known sessions by turn ID; dividers fall back to generic text for sessions not loaded yet. */
  sessions: Map<string, MessagingSession>;
  platform: ChannelPlatform | null;
  busy: boolean;
  /** Session to scroll to; `key` changes on every jump request. */
  focus: SessionFocus | null;
  /** Called once the history has scrolled to `focus`. */
  onFocused: () => void;
  onLoadOlder: () => void;
}) {
  const t = useT();
  const locale = useLocale();
  const list = useRef<HTMLOListElement>(null);
  const messages = [...detail.messages].reverse();
  const latest = detail.conversation.turn;
  const latestSession = latest ? sessions.get(latest.id) : undefined;

  const atBottom = useRef(true);
  const [unseen, setUnseen] = useState(false);
  // Messages are newest first, so the first id changes exactly when something newer arrives.
  const newestID = detail.messages[0]?.id;
  const seenNewest = useRef<string | undefined>(undefined);
  useEffect(() => {
    const element = list.current;
    const first = seenNewest.current === undefined;
    seenNewest.current = newestID;
    if (!element) return;
    if (first || atBottom.current) {
      element.scrollTop = element.scrollHeight;
      return;
    }
    setUnseen(true);
  }, [newestID]);
  function onScroll() {
    const element = list.current;
    if (!element) return;
    atBottom.current = element.scrollHeight - element.scrollTop - element.clientHeight < NEAR_BOTTOM_PX;
    if (atBottom.current) setUnseen(false);
  }

  // Declared after the bottom-scroll effect so a session jump wins over it on the same render.
  const focused = useEffectEvent(() => onFocused());
  useEffect(() => {
    if (!focus) return;
    const anchor = list.current?.querySelector<HTMLElement>(`#${CSS.escape(sessionAnchorID(focus.turnID))}`);
    if (anchor) {
      anchor.scrollIntoView({ block: "start", behavior: "smooth" });
      anchor.focus({ preventScroll: true });
      focused();
    }
  }, [focus, detail.messages]);

  function startLabel(session: MessagingSession | undefined) {
    if (!session) return t("desk.turnDivider.newTurn");
    const range = t("desk.sessions.dividerOpened", { range: sessionRange(t, session, locale) });
    return session.turn.lifecycle === "closed"
      ? `${range} · ${sessionCause(t, session.turn)} · ${sessionOutcome(t, session.turn)}`
      : range;
  }

  return (
    <>
      {detail.next_cursor && (
        <button type="button" className={`${secondaryClass} self-center`} disabled={busy} onClick={onLoadOlder}>
          {t("desk.loadOlder")}
        </button>
      )}
      <ol
        ref={list}
        onScroll={onScroll}
        className="max-h-[520px] space-y-3 overflow-y-auto"
        aria-label={t("desk.history")}
        aria-live="polite"
        aria-relevant="additions"
      >
        {messages.map((message, index) => {
          const session = sessions.get(message.turn_id);
          const boundary = index > 0 && messages[index - 1].turn_id !== message.turn_id;
          // The first loaded message starts a session only when it is that session's first message.
          const starts = boundary || (index === 0 && !!session && session.first_message_id === message.id);
          return (
            <Fragment key={message.id}>
              {starts && (
                <TurnDivider id={session?.first_message_id === message.id ? sessionAnchorID(message.turn_id) : undefined}>
                  {startLabel(session)}
                </TurnDivider>
              )}
              <MessageBubble
                message={message}
                customerName={detail.conversation.thread.display_name}
                platform={platform}
              />
            </Fragment>
          );
        })}
        {latest?.lifecycle === "closed" && (
          <TurnDivider>
            {latestSession
              ? t("desk.sessions.dividerClosed", { cause: sessionCause(t, latest), outcome: sessionOutcome(t, latest) })
              : t("desk.turnDivider.closed")}
          </TurnDivider>
        )}
      </ol>
      {unseen && (
        <button
          type="button"
          className={`${secondaryClass} self-center`}
          onClick={() => {
            const element = list.current;
            if (element) element.scrollTop = element.scrollHeight;
            setUnseen(false);
          }}
        >
          {t("desk.newMessages")}
        </button>
      )}
    </>
  );
}
