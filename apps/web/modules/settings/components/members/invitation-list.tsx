"use client";

import { EmptyState, StatusChip, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { Invitation } from "@/lib/services/onboarding";
import { inviteChip } from "../../utils";

interface Props {
  invites: Invitation[];
  busy: boolean;
  onRevoke: (id: string) => void;
}

/** Owner view of invitations; pending ones can be revoked. */
export function InvitationList({ invites, busy, onRevoke }: Props) {
  const t = useT();
  return (
    <section className="space-y-3 rounded-xl border border-border bg-surface p-5">
      <h2 className="font-semibold">{t("members.invitations")}</h2>
      {!invites.length ? (
        <EmptyState title={t("members.noInvitations")} />
      ) : (
        <ul className="divide-y divide-border">
          {invites.map((invite) => {
            const chip = inviteChip(t, invite.status);
            return (
              <li key={invite.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                <div>
                  <p className="text-sm font-medium">{invite.email}</p>
                  <p className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted">
                    <StatusChip tone={chip.tone}>{chip.label}</StatusChip>
                    {t("members.expires", { date: new Date(invite.expires_at).toLocaleString() })}
                  </p>
                </div>
                {invite.status === "pending" && (
                  <button type="button" className={secondaryClass} disabled={busy} onClick={() => onRevoke(invite.id)}>
                    {t("members.revoke")}
                  </button>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
