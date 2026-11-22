"use client";

import { useState } from "react";
import { Field, buttonClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { telegramTokenPattern } from "../../utils";
import { TelegramTokenInput } from "./telegram-token-input";

interface Props {
  busy: string;
  onConnect: (token: string) => Promise<unknown>;
}

/** Owner form to connect a bot. The raw token is cleared on submit whatever the outcome. */
export function TelegramConnectForm({ busy, onConnect }: Props) {
  const t = useT();
  const [token, setToken] = useState("");
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        if (busy || !telegramTokenPattern.test(token)) return;
        const submitted = token;
        setToken("");
        void onConnect(submitted);
      }}
      className="space-y-2"
    >
      <Field title={t("channels.telegram.token")}>
        <TelegramTokenInput id="telegram-token" value={token} disabled={!!busy} onChange={setToken} />
      </Field>
      <p id="telegram-token-help" className="text-xs text-muted">
        {t("channels.telegram.tokenHelp")}
      </p>
      <button className={buttonClass} disabled={!!busy || !telegramTokenPattern.test(token)}>
        {busy === "connect" ? t("channels.telegram.connecting") : t("channels.telegram.connect")}
      </button>
    </form>
  );
}
