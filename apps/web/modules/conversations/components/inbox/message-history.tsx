"use client";

import { useEffect, useRef, useState } from "react";
import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ChannelPlatform } from "@/modules/settings";
import type { MessagingDetail } from "../../services/messaging";
import { MessageBubble } from "./message-bubble";

const NEAR_BOTTOM_PX = 80;

/**
 * Messages oldest-first, with a button to page in older ones. A newly arrived message scrolls into view when the
 * reader is already at the bottom; otherwise a "new messages" button keeps their place.
 */
export function MessageHistory({
  detail,
  platform,
  busy,
  onLoadOlder,
}: {
  detail: MessagingDetail;
  platform: ChannelPlatform | null;
  busy: boolean;
  onLoadOlder: () => void;
}) {
  const t = useT();
  const list = useRef<HTMLOListElement>(null);
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
        {[...detail.messages].reverse().map((message) => (
          <MessageBubble
            key={message.id}
            message={message}
            customerName={detail.conversation.thread.display_name}
            platform={platform}
          />
        ))}
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
