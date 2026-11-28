"use client";

import { useT } from "@/lib/i18n/provider";
import { platformLabel, type ChannelPlatform } from "@/modules/settings";
import { botActorID, type MessagingMessage } from "../../services/messaging";
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
    message.direction === "outbound" && message.actor_id === botActorID
      ? message.status === "channel"
        ? t("desk.sender.bot", { channel })
        : t("desk.sender.botSimulated")
      : message.status === "received"
      ? `${t("desk.sender.customer")} · ${customerName}`
      : message.status === "channel"
        ? t("desk.sender.sale", { channel })
        : message.direction === "inbound"
          ? t("desk.sender.customerSimulated")
          : t("desk.sender.saleSimulated");
  const outbound = message.direction === "outbound";
  const bot = outbound && message.actor_id === botActorID;
  const tone = !outbound
    ? "mr-auto rounded-bl-sm border border-border bg-surface"
    : bot
      ? "ml-auto rounded-br-sm bg-primary/10"
      : "ml-auto rounded-br-sm bg-primary text-white [&_*]:text-white";
  return (
    <li className={`max-w-[85%] rounded-2xl px-3.5 py-2.5 shadow-sm ${tone}`}>
      <p className={`text-xs ${outbound && !bot ? "opacity-80" : "text-muted"}`}>
        {sender} · {new Date(message.created_at).toLocaleString()}
      </p>
      <MessageContent message={message} platform={platform} />
      <DeliveryStatus message={message} platform={platform} />
    </li>
  );
}
