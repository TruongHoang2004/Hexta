"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { platformLabel, type ChannelPlatform } from "@/modules/settings";
import type { MessagingMessage } from "../../services/messaging";
import { deliveryReasons, type DeliveryReason } from "../../types";
import { isOneOf } from "../../utils";

/** Whether a message reached the customer's channel. */
export function DeliveryStatus({
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
              place: platform === "telegram" ? t("desk.delivery.placeTelegram") : t("desk.delivery.placeFacebook"),
            })}
          </p>
        </div>
      );
    default:
      return <p className="mt-1 text-xs text-muted">{t("desk.delivery.sending")}</p>;
  }
}
