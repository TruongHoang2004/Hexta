"use client";

import { useState } from "react";
import { Field, buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { TelegramConnection } from "../../types";
import { botName, telegramTokenPattern } from "../../utils";
import { TelegramTokenInput } from "./telegram-token-input";

interface Props {
  row: TelegramConnection;
  busy: string;
  onRotate: (token: string) => Promise<unknown>;
  onCancel: () => void;
}

/** Replace a bot's token. The raw token is cleared on submit whatever the outcome. */
export function TelegramRotateForm({ row, busy, onRotate, onCancel }: Props) {
  const t = useT();
  const [token, setToken] = useState("");
  const id = `rotate-${row.id}`;
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        if (busy || !telegramTokenPattern.test(token)) return;
        const submitted = token;
        setToken("");
        void onRotate(submitted);
      }}
      className="space-y-2 rounded-lg border border-border p-3"
    >
      <Field title={t("channels.telegram.newToken", { name: botName(row) })}>
        <TelegramTokenInput id={id} value={token} disabled={!!busy} onChange={setToken} />
      </Field>
      <p id={`${id}-help`} className="text-xs text-muted">
        {t("channels.telegram.rotateHelp")}
      </p>
      <div className="flex gap-2">
        <button className={buttonClass} disabled={!!busy || !telegramTokenPattern.test(token)}>
          {busy === row.id ? t("channels.telegram.rotating") : t("channels.telegram.saveToken")}
        </button>
        <button type="button" className={secondaryClass} disabled={!!busy} onClick={onCancel}>
          {t("channels.cancel")}
        </button>
      </div>
    </form>
  );
}
