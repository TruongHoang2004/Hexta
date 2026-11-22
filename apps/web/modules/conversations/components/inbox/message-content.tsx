"use client";

import { useT } from "@/lib/i18n/provider";
import { platformLabel, type ChannelPlatform } from "@/modules/settings";
import type { MessagingMessage } from "../../services/messaging";
import { mediaKinds, type MediaKind } from "../../types";
import { isOneOf } from "../../utils";

/** Message text, or a placeholder for attachments the desk cannot show. */
export function MessageContent({
  message,
  platform,
}: {
  message: MessagingMessage;
  platform: ChannelPlatform | null;
}) {
  const t = useT();
  if (message.content_type !== "unsupported_media") {
    return <p className="mt-1 whitespace-pre-wrap break-words">{message.body}</p>;
  }
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
