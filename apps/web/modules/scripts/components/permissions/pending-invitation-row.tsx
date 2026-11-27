"use client";

import { useT } from "@/lib/i18n/provider";
import { scriptPermissions } from "../../types";

/** An invited person who has not joined yet: permissions can only be granted to current members. */
export function PendingInvitationRow({ email }: { email: string }) {
  const t = useT();
  return (
    <tr className="border-t border-border text-muted">
      <th scope="row" className="py-3 pr-3 text-left font-normal">
        <span className="block break-all font-medium">{email}</span>
        <span className="text-xs">{t("scriptPermissions.invitePending")}</span>
      </th>
      {scriptPermissions.map((permission) => (
        <td key={permission} className="px-2 py-3 text-center" aria-label={t("scriptPermissions.notAvailable")}>
          —
        </td>
      ))}
    </tr>
  );
}
