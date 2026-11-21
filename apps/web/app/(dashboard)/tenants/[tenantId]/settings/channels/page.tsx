"use client";

import { FacebookChannelSettings } from "@/components/channels/facebook-settings";
import { TelegramChannelSettings } from "@/components/channels/telegram-settings";
import { useT } from "@/lib/i18n/provider";

export default function ChannelsPage() {
  const t = useT();
  return (
    <div className="space-y-6">
      <p className="text-sm text-muted">{t("settingsPages.channelsIntro")}</p>
      <FacebookChannelSettings />
      <TelegramChannelSettings />
    </div>
  );
}
