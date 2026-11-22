"use client";

import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ChannelPlatform } from "@/modules/settings";
import type { MessagingDetail } from "../../services/messaging";
import { MessageBubble } from "./message-bubble";

/** Messages oldest-first, with a button to page in older ones. */
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
  return (
    <>
      {detail.next_cursor && (
        <button type="button" className={`${secondaryClass} self-center`} disabled={busy} onClick={onLoadOlder}>
          {t("desk.loadOlder")}
        </button>
      )}
      <ol className="max-h-[520px] space-y-3 overflow-y-auto" aria-label={t("desk.history")}>
        {[...detail.messages].reverse().map((message) => (
          <MessageBubble
            key={message.id}
            message={message}
            customerName={detail.conversation.thread.display_name}
            platform={platform}
          />
        ))}
      </ol>
    </>
  );
}
