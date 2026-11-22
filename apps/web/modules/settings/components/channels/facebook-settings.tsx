"use client";

import { Notice, buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useFacebookConnections } from "../../hooks/use-facebook-connections";
import { ExpiredConnectionBanner } from "./expired-connection-banner";
import { FacebookConnectionRow } from "./facebook-connection-row";

/** Facebook Pages connected to this workspace; owners connect and manage them. */
export function FacebookSettings() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const fb = useFacebookConnections(tenant?.id ?? "");
  if (!tenant) return null;
  const { rows, loading, busy } = fb;
  return (
    <div className="space-y-6">
      <ExpiredConnectionBanner tenant={tenant.id} connections={rows} owner={isOwner} />
      <Notice error={fb.error} success={fb.success} />
      <div className="flex flex-wrap gap-2">
        {isOwner && (
          <button className={buttonClass} disabled={!!busy} onClick={() => void fb.connect()}>
            {busy === "connect" ? t("channels.facebook.opening") : t("channels.facebook.connect")}
          </button>
        )}
        <button className={secondaryClass} disabled={loading || !!busy} onClick={fb.refresh}>
          {t("channels.refresh")}
        </button>
      </div>
      {!isOwner && <p className="text-sm text-muted">{t("channels.facebook.ownerOnly")}</p>}
      <section className="rounded-xl border border-border bg-surface p-5" aria-labelledby="facebook-pages">
        <h2 id="facebook-pages" className="font-semibold">
          {t("channels.facebook.title")}
        </h2>
        {loading ? (
          <p role="status" className="mt-3">
            {t("channels.facebook.loading")}
          </p>
        ) : !rows.length ? (
          <p className="mt-3 text-sm text-muted">{t("channels.facebook.empty")}</p>
        ) : (
          <ul className="mt-2 divide-y divide-border">
            {rows.map((row) => (
              <FacebookConnectionRow
                key={row.id}
                row={row}
                owner={isOwner}
                busy={busy}
                onReconnect={() => void fb.connect()}
                onAct={(action) => void fb.act(row, action)}
              />
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
