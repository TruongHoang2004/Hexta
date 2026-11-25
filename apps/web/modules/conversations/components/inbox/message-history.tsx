"use client";

import { Fragment } from "react";
import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ChannelPlatform } from "@/modules/settings";
import type { MessagingDetail } from "../../services/messaging";
import { MessageBubble } from "./message-bubble";
import { TurnDivider } from "./turn-divider";

/**
 * Messages oldest-first, with a button to page in older ones. A divider marks each turn boundary (turns only close
 * after 24 hours of inactivity) and the end of a closed current turn.
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
  const messages = [...detail.messages].reverse();
  const closed = detail.conversation.turn?.lifecycle === "closed";
  return (
    <>
      {detail.next_cursor && (
        <button type="button" className={`${secondaryClass} self-center`} disabled={busy} onClick={onLoadOlder}>
          {t("desk.loadOlder")}
        </button>
      )}
      <ol className="max-h-[520px] space-y-3 overflow-y-auto" aria-label={t("desk.history")}>
        {messages.map((message, index) => (
          <Fragment key={message.id}>
            {index > 0 && messages[index - 1].turn_id !== message.turn_id && (
              <TurnDivider>{t("desk.turnDivider.newTurn")}</TurnDivider>
            )}
            <MessageBubble
              message={message}
              customerName={detail.conversation.thread.display_name}
              platform={platform}
            />
          </Fragment>
        ))}
        {closed && <TurnDivider>{t("desk.turnDivider.closed")}</TurnDivider>}
      </ol>
    </>
  );
}
