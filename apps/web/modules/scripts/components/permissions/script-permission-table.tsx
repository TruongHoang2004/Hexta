"use client";

import type { Invitation } from "@/lib/services/onboarding";
import { useT } from "@/lib/i18n/provider";
import type { MemberScriptPermissions, ScriptPermission } from "../../types";
import { scriptPermissions } from "../../types";
import { permissionLabelKey } from "../../utils";
import { PendingInvitationRow } from "./pending-invitation-row";
import { ScriptPermissionRow } from "./script-permission-row";

interface Props {
  members: MemberScriptPermissions[];
  invites: Invitation[];
  nameOf: (member: MemberScriptPermissions) => string;
  pending: string;
  onChange: (member: MemberScriptPermissions, permission: ScriptPermission, granted: boolean) => void;
}

/** Member × permission grid (S-17). Owner first, then members, then pending invitations. */
export function ScriptPermissionTable({ members, invites, nameOf, pending, onChange }: Props) {
  const t = useT();
  const ordered = [...members].sort((a, b) => Number(b.owner) - Number(a.owner));
  return (
    <div className="overflow-x-auto rounded-xl border border-border bg-surface p-5">
      <table className="w-full min-w-[560px] text-sm">
        <caption className="sr-only">{t("scriptPermissions.title")}</caption>
        <thead>
          <tr className="text-xs text-muted">
            <th scope="col" className="pb-2 text-left font-medium">
              {t("scriptPermissions.member")}
            </th>
            {scriptPermissions.map((permission) => (
              <th key={permission} scope="col" className="px-2 pb-2 text-center font-medium">
                <span className="block text-foreground">{t(permissionLabelKey(permission))}</span>
                <code className="text-[11px]">{permission}</code>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {ordered.map((member) => (
            <ScriptPermissionRow key={member.user_id} member={member} name={nameOf(member)} pending={pending} onChange={onChange} />
          ))}
          {invites.map((invite) => (
            <PendingInvitationRow key={invite.id} email={invite.email} />
          ))}
        </tbody>
      </table>
    </div>
  );
}
