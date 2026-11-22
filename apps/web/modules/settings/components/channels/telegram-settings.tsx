"use client";

import { useState } from "react";
import { Notice, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useTelegramConnections } from "../../hooks/use-telegram-connections";
import { ExpiredConnectionBanner } from "./expired-connection-banner";
import { TelegramConnectForm } from "./telegram-connect-form";
import { TelegramConnectionRow } from "./telegram-connection-row";

/** Telegram bots connected to this workspace; owners connect, rotate tokens and manage them. */
export function TelegramSettings() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const bots = useTelegramConnections(tenant?.id ?? "");
  // The row whose rotate-token form is open; closing it unmounts the form and drops any typed token.
  const [rotating, setRotating] = useState("");
  if (!tenant) return null;
  const { rows, loading, busy } = bots;
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5" aria-labelledby="telegram-bots">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 id="telegram-bots" className="font-semibold">
          {t("channels.telegram.title")}
        </h2>
        <button className={secondaryClass} disabled={loading || !!busy} onClick={bots.refresh}>
          {t("channels.refresh")}
        </button>
      </div>
      <p className="text-sm text-muted">{t("channels.telegram.intro")}</p>
      <ExpiredConnectionBanner tenant={tenant.id} connections={rows} owner={isOwner} />
      <Notice error={bots.error} success={bots.success} />
      {isOwner ? (
        <TelegramConnectForm busy={busy} onConnect={bots.connect} />
      ) : (
        <p className="text-sm text-muted">{t("channels.telegram.ownerOnly")}</p>
      )}
      {loading ? (
        <p role="status">{t("channels.telegram.loading")}</p>
      ) : !rows.length ? (
        <p className="text-sm text-muted">{t("channels.telegram.empty")}</p>
      ) : (
        <ul className="divide-y divide-border">
          {rows.map((row) => (
            <TelegramConnectionRow
              key={row.id}
              row={row}
              owner={isOwner}
              busy={busy}
              rotating={rotating === row.id}
              statusText={bots.statusText}
              onToggleRotate={() => setRotating(rotating === row.id ? "" : row.id)}
              onRotate={async (token) => {
                if (await bots.rotate(row, token)) setRotating("");
              }}
              onAct={(action) => void bots.act(row, action)}
            />
          ))}
        </ul>
      )}
    </section>
  );
}
