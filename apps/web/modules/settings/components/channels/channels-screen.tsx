"use client";

import { useT } from "@/lib/i18n/provider";
import { FacebookSettings } from "./facebook-settings";
import { TelegramSettings } from "./telegram-settings";

export function ChannelsScreen() {
  const t = useT();
  return (
    <div className="space-y-6">
      <p className="text-sm text-muted">{t("settingsPages.channelsIntro")}</p>
      <FacebookSettings />
      <TelegramSettings />
    </div>
  );
}
