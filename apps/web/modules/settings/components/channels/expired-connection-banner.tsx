"use client";

import Link from "next/link";
import { useT } from "@/lib/i18n/provider";
import type { ChannelConnection } from "../../types";
import { platformLabel } from "../../utils";

// Shown in Inbox and Settings when any Page or bot needs an owner to reconnect or rotate its token. Sales see the
// status; owners get the action.
export function ExpiredConnectionBanner({
  tenant,
  connections,
  owner,
}: {
  tenant: string;
  connections: ChannelConnection[];
  owner: boolean;
}) {
  const t = useT();
  const expired = connections.filter((c) => c.status === "auth_expired");
  if (!expired.length) return null;
  const names = expired
    .map(
      (c) =>
        `${c.account_name || (c.account_username ? `@${c.account_username}` : c.account_id)} (${platformLabel[c.platform]})`,
    )
    .join(", ");
  return (
    <div role="alert" className="rounded-xl border border-error bg-error/10 p-3 text-sm">
      <p className="font-medium">{t("desk.expired.title", { names })}</p>
      <p className="mt-1">
        {t("desk.expired.body")}{" "}
        {owner ? (
          <Link href={`/tenants/${tenant}/settings/channels`} className="font-medium underline">
            {t("desk.expired.fix")}
          </Link>
        ) : (
          t("desk.expired.askOwner")
        )}
      </p>
    </div>
  );
}
