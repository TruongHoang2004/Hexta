"use client";

import { useT } from "@/lib/i18n/provider";
import { platformLabel, type ChannelPlatform } from "@/modules/settings";
import type { MessagingMessage } from "../../services/messaging";
import { DeliveryStatus } from "./delivery-status";
import { MessageContent } from "./message-content";

/** One message in the history: sender line, content and delivery state. */
export function MessageBubble({
  message,
  customerName,
  platform,
}: {
  message: MessagingMessage;
  customerName: string;
  platform: ChannelPlatform | null;
}) {
  const t = useT();
  const channel = platform ? platformLabel[platform] : "";
  const sender =
    message.status === "received"
      ? `${t("desk.sender.customer")} · ${customerName}`
      : message.status === "channel"
        ? t("desk.sender.sale", { channel })
        : message.direction === "inbound"
          ? t("desk.sender.customerSimulated")
          : t("desk.sender.saleSimulated");
  return (
    <li
      className={`max-w-[85%] rounded-xl border border-border p-3 ${message.direction === "outbound" ? "ml-auto bg-primary/5" : "mr-auto bg-surface"}`}
    >
      <p className="text-xs text-muted">
        {sender} · {new Date(message.created_at).toLocaleString()}
      </p>
      <MessageContent message={message} platform={platform} />
      <DeliveryStatus message={message} platform={platform} />
    </li>
  );
}
