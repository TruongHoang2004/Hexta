"use client";

import { ConflictBanner, EmptyState, Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import { useScriptPermissions } from "../../hooks/use-script-permissions";
import type { MemberScriptPermissions } from "../../types";
import { permissionLabelKey } from "../../utils";
import { ScriptPermissionTable } from "./script-permission-table";

/** S-17 Settings › Script permissions. Only the owner grants or revokes; changes apply on the member's next action. */
export function ScriptPermissionsScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const me = useAuthStore((state) => state.user?.id);
  const state = useScriptPermissions(tenant?.id ?? "", isOwner);
  const nameOf = (member: MemberScriptPermissions) =>
    member.user_id === me ? `${member.email || t("common.you")} (${t("common.you")})` : member.email || `#${member.user_id.slice(-6)}`;

  if (!isOwner) {
    return <p className="text-sm text-muted">{t("scriptPermissions.ownerOnly")}</p>;
  }
  const changed = state.changed;
  const success = changed
    ? t(changed.granted ? "scriptPermissions.granted" : "scriptPermissions.revoked", {
        permission: t(permissionLabelKey(changed.permission)),
        name: nameOf(changed.member),
      })
    : "";
  const onlyOwner = !state.loading && state.members.every((m) => m.owner) && !state.invites.length;
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold">{t("scriptPermissions.title")}</h2>
        <p className="mt-1 max-w-3xl text-sm text-muted">{t("scriptPermissions.intro")}</p>
      </div>
      {state.conflict ? (
        <ConflictBanner message={t("scriptPermissions.conflict")} actionLabel={t("scriptPermissions.reload")} onReload={state.reload} />
      ) : (
        <Notice error={state.error} success={success} />
      )}
      {state.loading ? (
        <p role="status">{t("scriptPermissions.loading")}</p>
      ) : (
        <>
          <ScriptPermissionTable
            members={state.members}
            invites={state.invites}
            nameOf={nameOf}
            pending={state.pending}
            onChange={(member, permission, granted) => void state.change(member, permission, granted)}
          />
          {onlyOwner && <EmptyState title={t("scriptPermissions.emptyTitle")} description={t("scriptPermissions.emptyBody")} />}
        </>
      )}
    </div>
  );
}
