"use client";

import { useState } from "react";
import { Notice, SegmentedFilter, buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useConnections } from "../../hooks/use-connections";
import type { ChannelPlatform } from "../../types";
import { channelPlatforms, platformLabel } from "../../utils";
import { AddConnectionDialog } from "./add-connection-dialog";
import { ConnectionRow } from "./connection-row";
import { ExpiredConnectionBanner } from "./expired-connection-banner";

/** All connections of the workspace in one list; a workspace may hold any number of each type. */
export function ChannelsScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const list = useConnections(tenant?.id ?? "");
  const [adding, setAdding] = useState(false);
  const [filter, setFilter] = useState<ChannelPlatform | "all">("all");
  // The row whose rotate-token form is open; closing it unmounts the form and drops any typed token.
  const [rotating, setRotating] = useState("");
  if (!tenant) return null;
  const { rows, loading, busy } = list;
  const visible = filter === "all" ? rows : rows.filter((row) => row.platform === filter);
  return (
    <div className="space-y-6">
      <p className="text-sm text-muted">{t("settingsPages.channelsIntro")}</p>
      <ExpiredConnectionBanner tenant={tenant.id} connections={rows} owner={isOwner} />
      <Notice error={list.error} success={list.success} />
      <div className="flex flex-wrap items-center gap-2">
        {isOwner && (
          <button className={buttonClass} disabled={!!busy} onClick={() => setAdding(true)}>
            {t("channels.addConnection")}
          </button>
        )}
        <button className={secondaryClass} disabled={loading || !!busy} onClick={list.refresh}>
          {t("channels.refresh")}
        </button>
      </div>
      {!isOwner && <p className="text-sm text-muted">{t("channels.ownerOnly")}</p>}
      <section className="space-y-3 rounded-xl border border-border bg-surface p-5" aria-labelledby="connections">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 id="connections" className="font-semibold">
            {t("channels.listTitle")}
          </h2>
          <SegmentedFilter
            label={t("channels.filterLabel")}
            value={filter}
            onChange={setFilter}
            options={[
              { value: "all", label: t("channels.all"), count: rows.length },
              ...channelPlatforms.map((platform) => ({
                value: platform,
                label: platformLabel[platform],
                count: rows.filter((row) => row.platform === platform).length,
              })),
            ]}
          />
        </div>
        {loading ? (
          <p role="status">{t("channels.loading")}</p>
        ) : !visible.length ? (
          <p className="text-sm text-muted">{t("channels.empty")}</p>
        ) : (
          <ul className="divide-y divide-border">
            {visible.map((row) => (
              <ConnectionRow
                key={row.id}
                row={row}
                owner={isOwner}
                busy={busy}
                statusText={list.statusText(row)}
                rotating={rotating === row.id}
                onToggleRotate={() => setRotating(rotating === row.id ? "" : row.id)}
                onRotate={async (token) => {
                  if (await list.rotate(row, token)) setRotating("");
                }}
                onReconnect={() => void list.connectFacebook()}
                onAct={(action) => void list.act(row, action)}
              />
            ))}
          </ul>
        )}
      </section>
      {isOwner && (
        <AddConnectionDialog
          open={adding}
          busy={busy}
          onClose={() => setAdding(false)}
          onFacebook={() => void list.connectFacebook()}
          onTelegram={list.connectTelegram}
        />
      )}
    </div>
  );
}
