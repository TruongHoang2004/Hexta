"use client";

import Link from "next/link";
import type { FacebookConnection } from "@/lib/services/channels";

// Shown in Inbox and Settings when any Page needs an owner to reconnect. Sales see the status; owners get the action.
export function ExpiredConnectionBanner({
  tenant,
  connections,
  owner,
}: {
  tenant: string;
  connections: FacebookConnection[];
  owner: boolean;
}) {
  const expired = connections.filter((c) => c.status === "auth_expired");
  if (!expired.length) return null;
  return (
    <div
      role="alert"
      className="rounded-xl border border-red-300 bg-red-50 p-3 text-sm text-red-900"
    >
      <p className="font-medium">
        Facebook authorization expired for{" "}
        {expired.map((c) => c.account_name || c.account_id).join(", ")}.
      </p>
      <p className="mt-1">
        Replies to these Pages cannot be sent until the Page is reconnected.{" "}
        {owner ? (
          <Link
            href={`/tenants/${tenant}/settings/channels`}
            className="font-medium underline"
          >
            Reconnect in Channel settings
          </Link>
        ) : (
          "Ask the workspace owner to reconnect it."
        )}
      </p>
    </div>
  );
}
