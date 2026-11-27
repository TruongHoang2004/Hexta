"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@hexta/ui";
import { buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ChannelPlatform } from "../../types";
import { channelPlatforms } from "../../utils";
import { TelegramConnectForm } from "./telegram-connect-form";

interface Props {
  open: boolean;
  busy: string;
  onClose: () => void;
  onFacebook: () => void;
  onTelegram: (token: string) => Promise<boolean>;
}

/** Owner picks a connection type, then completes that type's own connect flow. */
export function AddConnectionDialog({ open, busy, onClose, onFacebook, onTelegram }: Props) {
  const t = useT();
  const [type, setType] = useState<ChannelPlatform | "">("");
  const close = () => {
    setType("");
    onClose();
  };
  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next && !busy) close();
      }}
    >
      <DialogContent>
        <DialogTitle>{t("channels.addConnection")}</DialogTitle>
        <DialogDescription>{type ? t(`channels.types.${type}.description`) : t("channels.chooseType")}</DialogDescription>
        {!type ? (
          <ul className="space-y-2">
            {channelPlatforms.map((platform) => (
              <li key={platform}>
                <button
                  type="button"
                  className="w-full rounded-xl border border-border p-4 text-left hover:bg-surface-hover focus-visible:outline focus-visible:outline-2"
                  onClick={() => setType(platform)}
                >
                  <span className="block font-medium">{t(`channels.types.${platform}.title`)}</span>
                  <span className="mt-1 block text-sm text-muted">{t(`channels.types.${platform}.description`)}</span>
                </button>
              </li>
            ))}
          </ul>
        ) : type === "facebook" ? (
          <div className="space-y-3">
            <p className="text-sm text-muted">{t("channels.facebook.oauthHelp")}</p>
            <button className={buttonClass} disabled={!!busy} onClick={onFacebook}>
              {busy === "connect" ? t("channels.facebook.opening") : t("channels.facebook.connect")}
            </button>
          </div>
        ) : (
          <TelegramConnectForm
            busy={busy}
            onConnect={async (token) => {
              if (await onTelegram(token)) close();
            }}
          />
        )}
        {type && (
          <button type="button" className={secondaryClass} disabled={!!busy} onClick={() => setType("")}>
            {t("channels.back")}
          </button>
        )}
      </DialogContent>
    </Dialog>
  );
}
