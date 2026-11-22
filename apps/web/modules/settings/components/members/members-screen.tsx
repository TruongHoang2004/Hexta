"use client";

import { Notice, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useMembers } from "../../hooks/use-members";
import { InvitationList } from "./invitation-list";
import { InviteForm } from "./invite-form";
import { MemberList } from "./member-list";

/** Team members; owners also invite people and manage pending invitations. */
export function MembersScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const team = useMembers(tenant?.id ?? "", isOwner);
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted">{t("members.intro")}</p>
        <button type="button" className={secondaryClass} disabled={team.loading || team.busy} onClick={team.refresh}>
          {t("members.refresh")}
        </button>
      </div>
      <Notice error={team.error} />
      {isOwner && <InviteForm busy={team.busy} link={team.link} onInvite={team.invite} />}
      <MemberList members={team.members} loading={team.loading} />
      {isOwner && <InvitationList invites={team.invites} busy={team.busy} onRevoke={(id) => void team.revoke(id)} />}
    </div>
  );
}
