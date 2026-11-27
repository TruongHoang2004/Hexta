"use client";

import { Checkbox } from "@hexta/ui";
import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { MemberScriptPermissions, ScriptPermission } from "../../types";
import { scriptPermissions } from "../../types";
import { permissionLabelKey } from "../../utils";

interface Props {
  member: MemberScriptPermissions;
  name: string;
  /** "userID:permission" of the change in flight, if any. */
  pending: string;
  onChange: (member: MemberScriptPermissions, permission: ScriptPermission, granted: boolean) => void;
}

/** One member × four permission toggles. The owner row is locked: ownership always implies every permission. */
export function ScriptPermissionRow({ member, name, pending, onChange }: Props) {
  const t = useT();
  return (
    <tr className="border-t border-border">
      <th scope="row" className="py-3 pr-3 text-left font-normal">
        <span className="block break-all font-medium">{name}</span>
        <span className="text-xs text-muted">
          {member.owner ? t("scriptPermissions.ownerRow") : t("roles.member")}
        </span>
      </th>
      {scriptPermissions.map((permission) => {
        const label = t("scriptPermissions.toggleLabel", { permission: t(permissionLabelKey(permission)), name });
        return (
          <td key={permission} className="px-2 py-3 text-center">
            {member.owner ? (
              <StatusChip tone="solid">{t("scriptPermissions.always")}</StatusChip>
            ) : (
              <Checkbox
                aria-label={label}
                checked={member.permissions.includes(permission)}
                disabled={pending !== ""}
                aria-busy={pending === `${member.user_id}:${permission}`}
                onCheckedChange={(checked) => onChange(member, permission, checked === true)}
              />
            )}
          </td>
        );
      })}
    </tr>
  );
}
