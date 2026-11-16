"use client";

import Link from "next/link";
import {
  platformLabel,
  type ChannelConnection,
} from "@/lib/services/channels";

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
  const expired = connections.filter((c) => c.status === "auth_expired");
  if (!expired.length) return null;
  const names = expired
    .map(
      (c) =>
        `${c.account_name || (c.account_username ? `@${c.account_username}` : c.account_id)} (${platformLabel[c.platform]})`,
    )
    .join(", ");
  return (
    <div
      role="alert"
      className="rounded-xl border border-red-300 bg-red-50 p-3 text-sm text-red-900"
    >
      <p className="font-medium">Channel authorization expired for {names}.</p>
      <p className="mt-1">
        Replies through these channels cannot be sent until a Facebook Page is
        reconnected or a Telegram bot token is rotated.{" "}
        {owner ? (
          <Link
            href={`/tenants/${tenant}/settings/channels`}
            className="font-medium underline"
          >
            Fix in Channel settings
          </Link>
        ) : (
          "Ask the workspace owner to fix it."
        )}
      </p>
    </div>
  );
}
